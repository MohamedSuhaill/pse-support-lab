// Package dataset defines the fictional customer files used by the lab.
package dataset

import (
	"fmt"
	"math/rand/v2"

	"invoicely/tools/internal/workbook"
)

type Party = workbook.Party

var (
	DesertRose   = Party{"Desert Rose Trading LLC", "100234567800003", "Office 1204, Al Fahidi Tower", "Dubai", "DXB"}
	OasisBuilder = Party{"Oasis Builders Supply LLC", "100318274600003", "Plot 17, Industrial Area 6", "Sharjah", "SHJ"}
	MarinaBay    = Party{"Marina Bay Hospitality LLC", "100427193500003", "Marina Walk, Building 3", "Dubai", "DXB"}
	AlNoor       = Party{"Al Noor Fuel & Lubricants LLC", "100536284100003", "Mussafah M-14, Street 9", "Abu Dhabi", "AUH"}
	Sahara       = Party{"Sahara Medical Supplies FZE", "100645371900003", "Warehouse B7, Ajman Free Zone", "Ajman", "AJM"}
	GhafCafe     = Party{"Ghaf Tree Café LLC", "100754862300003", "Corniche Road, Shop 4", "Fujairah", "FUJ"}
	PalmCoast    = Party{"Palm Coast Electronics LLC", "100863917200003", "Al Quoz 3, Warehouse 21", "Dubai", "DXB"}
	ZayedPort    = Party{"Zayed Port Logistics LLC", "100972845600003", "Mina Zayed, Gate 2", "Abu Dhabi", "AUH"}
)

var (
	PalmHospital = Party{"Emirates Palm Hospital LLC", "100118273600003", "Healthcare City, Building 42", "Dubai", "DXB"}
	AlWaha       = Party{"Al Waha Hypermarkets LLC", "100129384700003", "Al Wahda Street", "Sharjah", "SHJ"}
	Bluewater    = Party{"Bluewater Facilities Management LLC", "100130495800003", "Khalifa City A, Villa 18", "Abu Dhabi", "AUH"}
	Crescent     = Party{"Crescent Contracting LLC", "100141506900003", "Al Barsha 1, Office 302", "Dubai", "DXB"}
	GulfHorizon  = Party{"Gulf Horizon Transport LLC", "100152617000003", "Al Jurf Industrial 2", "Ajman", "AJM"}
	SaffronEvent = Party{"Saffron Events LLC", "100163728100003", "Dibba Road, Unit 9", "Fujairah", "FUJ"}
	Northstar    = Party{"Northstar Engineering Consultants LLC", "100174839200003", "Al Nahda 2, Office 11", "Dubai", "DXB"}
	CoralReef    = Party{"Coral Reef Resorts LLC", "100185940300003", "Al Marjan Island", "Ras Al Khaimah", "RAK"}
)

const iban = "AE070331234567890123456"

func S(item string, qty float64, unit string, price float64, itemType string) workbook.Line {
	return workbook.Line{Item: item, Qty: qty, Unit: unit, Price: price, Category: "S", ItemType: itemType}
}

// Sample is one named customer file.
type Sample struct {
	Key      string // short name used by `make replay CUSTOMER=<key>`
	Filename string
	SellerID int
	Build    func(suffix string) workbook.File
}

// Samples are the customer files shipped in samples/.
var Samples = []Sample{
	{"desert-rose", "desert-rose-nov-2026.xlsx", 1, desertRose},
	{"oasis", "oasis-builders-nov-2026.xlsx", 2, oasis},
	{"al-noor", "al-noor-fuel-nov-2026.xlsx", 4, alNoor},
	{"sahara", "sahara-medical-batch-1116.xlsx", 5, sahara},
	{"ghaf", "ghaf-cafe-nov-2026.xlsx", 6, ghaf},
}

func desertRose(suffix string) workbook.File {
	n := func(i int) string { return fmt.Sprintf("DRT-2611-%03d%s", i, suffix) }
	return workbook.File{
		Source: "Desert Rose Back Office", Seller: DesertRose, PaymentMeans: "30", Account: iban,
		Invoices: []workbook.Invoice{
			{n(1), "2026-11-10", "2026-12-10", AlWaha, []workbook.Line{
				S("Arabic coffee beans 1 kg", 40, "KGM", 85, "Goods"),
				S("Medjool dates gift box", 25, "BX", 120, "Goods"),
			}},
			{n(2), "2026-11-10", "2026-12-10", Bluewater, []workbook.Line{
				S("Pantry restocking service", 8, "HUR", 75, "Services"),
				S("Bottled water 24 x 500 ml", 60, "CT", 22, "Goods"),
				S("Green tea 100 bags", 30, "BX", 18, "Goods"),
			}},
			{n(3), "2026-11-11", "2026-12-11", PalmHospital, []workbook.Line{
				S("Medjool dates gift box", 50, "BX", 120, "Goods"),
				S("Saffron 10 g", 12, "PK", 95, "Goods"),
			}},
			{n(4), "2026-11-12", "2026-12-12", Northstar, []workbook.Line{
				S("Arabic coffee beans 1 kg", 15, "KGM", 85, "Goods"),
				S("Coffee machine servicing", 2, "HUR", 250, "Services"),
				S("Paper cups 1000 pcs", 6, "CT", 64, "Goods"),
			}},
			{n(5), "2026-11-13", "2026-12-13", AlWaha, []workbook.Line{
				S("Mixed nuts 1 kg", 35, "KGM", 72, "Goods"),
				S("Rose water 250 ml", 48, "EA", 14, "Goods"),
			}},
		},
	}
}

var buildingItems = []struct {
	name  string
	unit  string
	price float64
}{
	{"Portland cement 50 kg bag", "BG", 18}, {"Rebar 12 mm grade 60", "TNE", 2650}, {"Rebar 16 mm grade 60", "TNE", 2690},
	{"Concrete block 20 cm", "EA", 4}, {"Plywood sheet 18 mm", "EA", 96}, {"Ready-mix concrete C40", "MTQ", 310},
	{"Washed sand", "TNE", 38}, {"Aggregate 20 mm", "TNE", 42}, {"Binding wire 1.5 mm", "RO", 145},
	{"PVC pipe 110 mm 6 m", "EA", 54}, {"Waterproofing membrane 4 mm", "RO", 165}, {"Gypsum board 12.5 mm", "EA", 27},
	{"Tile adhesive 20 kg", "BG", 31}, {"Ceramic floor tile 60 x 60", "MTK", 48}, {"Copper cable 4 mm2 100 m", "RO", 385},
	{"Steel scaffolding tube 6 m", "EA", 120}, {"Safety helmet", "EA", 25}, {"Site lighting tower rental", "DAY", 220},
	{"Formwork carpentry", "HUR", 65}, {"Crane hire with operator", "HUR", 450},
}

func oasis(suffix string) workbook.File {
	sites := []Party{Crescent, Northstar, CoralReef, Bluewater, Crescent, GulfHorizon, Northstar, CoralReef}
	r := rand.New(rand.NewPCG(2611, 2))
	f := workbook.File{Source: "Oasis ERP Export", Seller: OasisBuilder, PaymentMeans: "30", Account: "AE450260001012345678901"}
	for i, site := range sites {
		inv := workbook.Invoice{
			Number:    fmt.Sprintf("OBS-2611-%03d%s", i+1, suffix),
			IssueDate: "2026-11-15", DueDate: "2026-12-15", Buyer: site,
		}
		for range 75 {
			it := buildingItems[r.IntN(len(buildingItems))]
			kind := "Goods"
			if it.unit == "HUR" || it.unit == "DAY" {
				kind = "Services"
			}
			inv.Lines = append(inv.Lines, S(it.name, float64(1+r.IntN(40)), it.unit, it.price, kind))
		}
		f.Invoices = append(f.Invoices, inv)
	}
	return f
}

func alNoor(suffix string) workbook.File {
	n := func(i int) string { return fmt.Sprintf("ANF-2611-%04d%s", i, suffix) }
	diesel := func(q, p float64) workbook.Line { return S("Diesel EN590", q, "LTR", p, "Goods") }
	return workbook.File{
		Source: "AlNoor FuelDesk", Seller: AlNoor, PaymentMeans: "30", Account: "AE210400000123456789012",
		Invoices: []workbook.Invoice{
			{n(41), "2026-11-12", "2026-12-12", GulfHorizon, []workbook.Line{diesel(4850, 2.835)}},
			{n(42), "2026-11-12", "2026-12-12", ZayedPort, []workbook.Line{
				diesel(3125, 2.917),
				S("Petrol Special 95", 1837, "LTR", 2.585, "Goods"),
				S("Engine oil 15W-40 5 L", 24, "EA", 18.45, "Goods"),
			}},
			{n(43), "2026-11-13", "2026-12-13", Crescent, []workbook.Line{
				diesel(6200, 2.835),
				S("Engine oil 15W-40 5 L", 60, "EA", 18.45, "Goods"),
			}},
			{n(44), "2026-11-13", "2026-12-13", GulfHorizon, []workbook.Line{
				diesel(2275, 2.917),
				S("Petrol Special 95", 1365, "LTR", 2.585, "Goods"),
				S("AdBlue", 415, "LTR", 1.275, "Goods"),
				S("Hydraulic oil 20 L", 12, "EA", 46.75, "Goods"),
			}},
			{n(45), "2026-11-14", "2026-12-14", Bluewater, []workbook.Line{diesel(3980, 2.835)}},
			{n(46), "2026-11-14", "2026-12-14", ZayedPort, []workbook.Line{
				diesel(1512, 2.917),
				S("AdBlue", 732, "LTR", 1.275, "Goods"),
			}},
		},
	}
}

var medicalItems = []struct {
	name  string
	unit  string
	price float64
}{
	{"Nitrile examination gloves (box of 100)", "BX", 45}, {"Disposable syringes 5 ml (box of 100)", "BX", 60},
	{"Surgical face masks (box of 50)", "BX", 30}, {"Sterile gauze swabs (pack of 100)", "PK", 26},
	{"IV cannula 20G (box of 50)", "BX", 140}, {"Alcohol prep pads (box of 200)", "BX", 22},
}

func sahara(suffix string) workbook.File {
	f := workbook.File{Source: "Bluepeak Connector", Seller: Sahara, PaymentMeans: "30", Account: "AE600500000000123456789"}
	for i := range 40 {
		buyer := PalmHospital
		if i%8 == 7 {
			buyer = CoralReef
		}
		inv := workbook.Invoice{
			Number:    fmt.Sprintf("SMS-1116-%03d%s", i+1, suffix),
			IssueDate: "2026-11-16", DueDate: "2026-12-16", Buyer: buyer,
		}
		a := medicalItems[i%len(medicalItems)]
		inv.Lines = append(inv.Lines, S(a.name, float64(10+i%7*5), a.unit, a.price, "Goods"))
		if i%2 == 1 {
			b := medicalItems[(i+3)%len(medicalItems)]
			inv.Lines = append(inv.Lines, S(b.name, float64(5+i%4*5), b.unit, b.price, "Goods"))
		}
		f.Invoices = append(f.Invoices, inv)
	}
	return f
}

func ghaf(suffix string) workbook.File {
	n := func(i int) string { return fmt.Sprintf("GTC-%03d%s", i, suffix) }
	return workbook.File{
		Source: "Manual", Seller: GhafCafe, PaymentMeans: "10",
		Invoices: []workbook.Invoice{
			{n(118), "13/11/2026", "13/12/2026", SaffronEvent, []workbook.Line{
				S("Corporate breakfast catering (per guest)", 40, "EA", 45, "Services"),
				S("Karak tea urn", 4, "EA", 60, "Goods"),
			}},
			{n(119), "13/11/2026", "13/12/2026", Northstar, []workbook.Line{
				S("Office lunch boxes", 25, "EA", 38, "Goods"),
				S("Delivery charge", 1, "EA", 30, "Services"),
			}},
			{n(120), "14/11/2026", "14/12/2026", CoralReef, []workbook.Line{
				S("Arabic sweets platter", 6, "EA", 85, "Goods"),
				S("Fresh juice jug 1.5 L", 10, "EA", 28, "Goods"),
				S("Event staff (per hour)", 8, "HUR", 40, "Services"),
			}},
			{n(121), "15/11/2026", "15/12/2026", SaffronEvent, []workbook.Line{
				S("Afternoon tea catering (per guest)", 60, "EA", 35, "Services"),
			}},
		},
	}
}

// PeakFile is a typical small upload from a busy customer during the morning peak.
func PeakFile(seller Party, prefix string, r *rand.Rand) workbook.File {
	buyers := []Party{PalmHospital, AlWaha, Bluewater, Crescent, GulfHorizon, Northstar, CoralReef}
	f := workbook.File{Source: "Peak upload", Seller: seller, PaymentMeans: "30", Account: iban}
	for i := range 6 {
		inv := workbook.Invoice{
			Number:    fmt.Sprintf("%s-%02d", prefix, i+1),
			IssueDate: "2026-11-16", DueDate: "2026-12-16", Buyer: buyers[r.IntN(len(buyers))],
		}
		for range 2 {
			it := buildingItems[r.IntN(len(buildingItems))]
			inv.Lines = append(inv.Lines, S(it.name, float64(1+r.IntN(20)), it.unit, it.price, "Goods"))
		}
		f.Invoices = append(f.Invoices, inv)
	}
	return f
}

// PeakSellers are the customers that upload during the morning peak.
var PeakSellers = []struct {
	ID     int
	Party  Party
	Prefix string
}{
	{1, DesertRose, "DRT"}, {3, MarinaBay, "MBH"}, {7, PalmCoast, "PCE"}, {8, ZayedPort, "ZPL"},
}
