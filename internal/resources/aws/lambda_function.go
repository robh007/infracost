package aws

import (
	"context"
	"math"
	"strconv"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/infracost/infracost/internal/usage/aws"

	"github.com/shopspring/decimal"
)

type LambdaFunction struct {
	Address      string
	Region       string
	Name         string
	MemorySize   int64
	Architecture string

	RequestDurationMS *int64 `infracost_usage:"request_duration_ms"`
	MonthlyRequests   *int64 `infracost_usage:"monthly_requests"`
}

func (a *LambdaFunction) CoreType() string {
	return "LambdaFunction"
}

func (a *LambdaFunction) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "request_duration_ms", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_requests", DefaultValue: 0, ValueType: schema.Int64},
	}
}

func (a *LambdaFunction) UsageEstimationParams() []schema.UsageParam {
	return []schema.UsageParam{
		{Key: "memory_size_gb", Value: decimal.NewFromInt(a.MemorySize).Div(decimal.NewFromInt(1024)).String()},
	}
}

func (a *LambdaFunction) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *LambdaFunction) BuildResource() *schema.Resource {
	memorySize := decimal.NewFromInt(a.MemorySize)
	requestType := "AWS-Lambda-Requests"
	durationType := "AWS-Lambda-Duration"

	averageRequestDuration := decimal.NewFromInt(1)
	if a.RequestDurationMS != nil {
		averageRequestDuration = decimal.NewFromInt(*a.RequestDurationMS)
	}

	if a.Architecture == "arm64" {
		requestType = "AWS-Lambda-Requests-ARM"
		durationType = "AWS-Lambda-Duration-ARM"
	}

	var monthlyRequests *decimal.Decimal
	var gbSeconds *decimal.Decimal
	var costComponents []*schema.CostComponent

	if a.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*a.MonthlyRequests))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: monthlyRequests,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AWSLambda"),
			ProductFamily: strPtr("Serverless"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr(requestType)},
				{Key: "usagetype", ValueRegex: strPtr("/Request/")},
			},
		},
	},
	)

	// Defaults to x86 tiers
	gbRequestTiers := []int{6000000000, 9000000000, 15000000000}
	displayNameList := []string{"Duration (first 6B)", "Duration (next 9B)", "Duration (over 15B)"}

	if a.Architecture == "arm64" {
		gbRequestTiers = []int{7500000000, 11250000000, 18750000000}
		displayNameList = []string{"Duration (first 7.5B)", "Duration (next 11.25B)", "Duration (over 18.75B)"}
	}

	if a.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*a.MonthlyRequests))
		gbSeconds = decimalPtr(calculateGBSeconds(memorySize, averageRequestDuration, *monthlyRequests))

		gbSecondQuantities := usage.CalculateTierBuckets(*gbSeconds, gbRequestTiers)

		costComponents = append(costComponents, a.durationCostComponent(displayNameList[0], "0", &gbSecondQuantities[0], durationType))

		if gbSecondQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, a.durationCostComponent(displayNameList[1], strconv.Itoa(gbRequestTiers[0]), &gbSecondQuantities[1], durationType))
		}

		if gbSecondQuantities[2].GreaterThanOrEqual(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, a.durationCostComponent(displayNameList[2], strconv.Itoa(gbRequestTiers[2]), &gbSecondQuantities[2], durationType))
		}

	} else {
		costComponents = append(costComponents, a.durationCostComponent(displayNameList[0], "0", gbSeconds, durationType))
	}

	estimate := func(ctx context.Context, values map[string]interface{}) error {
		inv, err := aws.LambdaGetInvocations(ctx, a.Region, a.Name)
		if err != nil {
			return err
		}
		values["monthly_requests"] = int64(math.Round(inv))
		dur, err := aws.LambdaGetDurationAvg(ctx, a.Region, a.Name)
		if err != nil {
			return err
		}
		values["request_duration_ms"] = int64(math.Round(dur))
		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: costComponents,
		EstimateUsage:  estimate,
	}
}

func calculateGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	gb := memorySize.Div(decimal.NewFromInt(1024))
	seconds := averageRequestDuration.Ceil().Div(decimal.NewFromInt(1000)) // Round up to closest 1ms and convert to seconds
	return monthlyRequests.Mul(gb).Mul(seconds)
}

func (a *LambdaFunction) durationCostComponent(displayName string, usageTier string, quantity *decimal.Decimal, durationType string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "GB-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AWSLambda"),
			ProductFamily: strPtr("Serverless"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr(durationType)},
				{Key: "usagetype", ValueRegex: strPtr("/GB-Second/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
