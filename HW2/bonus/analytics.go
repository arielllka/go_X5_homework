package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func AnalyzeSales(sales []Sale) Report {

    productUnits := make(map[string]int)
	report := Report{ RevenueByCategory: make(map[string]int64) }

	check:= func(sale Sale)bool{

		return (len(sale.Product) != 0) &&
		(len(sale.Category) !=0) && (sale.Quantity > 0) && 
		(sale.UnitPriceKopecks > 0) 

	}
		for _ , sale := range sales{

			if sale.Cancelled {
        		continue
            }

			if(check(sale)){

					RevenueKopecksUnit := int64(sale.Quantity) * sale.UnitPriceKopecks
					report.RevenueKopecks += RevenueKopecksUnit
					report.RevenueByCategory[sale.Category] += RevenueKopecksUnit
					report.UnitsSold += sale.Quantity
					productUnits[sale.Product] += sale.Quantity



			}else{
				report.InvalidSales ++
			}

		} 
		most_sale:=0

		for product , quantity := range productUnits{
			if (quantity > most_sale){
				report.TopProduct = product
				most_sale = quantity

			}else if(quantity == most_sale && product < report.TopProduct){
			 	report.TopProduct = product	
			}

		}

	
	return report
}
