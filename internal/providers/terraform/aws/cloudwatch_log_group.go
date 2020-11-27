package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudwatchLogGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_log_group",
		RFunc: NewCloudwatchLogGroup,
	}
}

func NewCloudwatchLogGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataIngestion *decimal.Decimal
	var gbDataStorage *decimal.Decimal
	var gbDataAnalyzed *decimal.Decimal
	var gbDataAnalyzedQueries int64

	if u != nil && u.Get("data_ingestion.0.value").Exists() {
		gbDataIngestion = decimalPtr(decimal.NewFromFloat(u.Get("data_ingestion.0.value").Float()))
	}

	if u != nil && u.Get("data_storage.0.value").Exists() {
		gbDataStorage = decimalPtr(decimal.NewFromFloat(u.Get("data_storage.0.value").Float()))
	}

	if u != nil && u.Get("data_analyzed.0.value").Exists() {
		gbDataAnalyzed = decimalPtr(decimal.NewFromFloat(u.Get("data_analyzed.0.value").Float()))
	}

	gbDataAnalyzedQueries = 1

	if u != nil && u.Get("data_queries.0.value").Exists() {
		gbDataAnalyzedQueries = u.Get("data_queries.0.value").Int()
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Data ingestion",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbDataIngestion,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataProcessing-Bytes/")},
					},
				},
			},
			{
				Name:            "Archival Storage",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbDataStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Storage Snapshot"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-ByteHrs/")},
					},
				},
			},
			{
				Name:            "Insights queries data scanned",
				Unit:            "GB",
				UnitMultiplier:  int(gbDataAnalyzedQueries),
				MonthlyQuantity: gbDataAnalyzed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataScanned-Bytes/")},
					},
				},
			},
		},
	}
}
