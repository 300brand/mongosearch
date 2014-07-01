package mongosearch

import (
	"time"
)

var today = time.Now().Format(TimeLayout)

var tests = []struct {
	Input     string
	Reduced   string
	MapReduce bool
}{
	{
		`keywords:a`,
		`keywords:a`,
		false,
	},
	{
		`keywords:(a OR "b c")`,
		`keywords:a keywords:"b c"`,
		true,
	},
	{
		`keywords:(a OR b)`,
		`keywords:a keywords:b`,
		false,
	},
	{
		`(keywords:(a OR b))`,
		`keywords:a keywords:b`,
		false,
	},
	{
		`(keywords:((a) OR (b)))`,
		`keywords:a keywords:b`,
		false,
	},
	{
		`(keywords:((a) AND (b)))`,
		`+keywords:a +keywords:b`,
		false,
	},
	{
		`published:(2014-01-01) AND keywords:(a AND b)`,
		`+keywords:a +keywords:b`,
		false,
	},
	{
		`(published:(2014-01-01) AND keywords:(a AND b))`,
		`+keywords:a +keywords:b`,
		false,
	},
	{
		// Search._id: "53a803bffc16f879a2000001"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:((("GDIT" AND "data center") OR ("General Dynamics Information Technology" AND "data center")))`,
		`(+keywords:"GDIT" +keywords:"data center") (+keywords:"General Dynamics Information Technology" +keywords:"data center")`,
		true,
	},
	{
		// Search._id: "53a804e8fc16f879a2000003"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:((("data center")))`,
		`+keywords:"data center"`,
		true,
	},
	{
		// Search._id: "53a8086bfc16f87dc1000003"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16' OR '2014-06-15' OR '2014-06-14' OR '2014-06-13' OR '2014-06-12' OR '2014-06-11' OR '2014-06-10' OR '2014-06-09' OR '2014-06-08' OR '2014-06-07' OR '2014-06-06' OR '2014-06-05' OR '2014-06-04' OR '2014-06-03' OR '2014-06-02' OR '2014-06-01' OR '2014-05-31' OR '2014-05-30' OR '2014-05-29' OR '2014-05-28' OR '2014-05-27' OR '2014-05-26' OR '2014-05-25' OR '2014-05-24') AND keywords:((("Continuous Monitoring")))`,
		`+keywords:"Continuous Monitoring"`,
		true,
	},
	{
		// Search._id: "53a81233fc16f8060f000003"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:((("Google" AND "data center") OR ("Facebook" AND "data center")))`,
		`(+keywords:"Google" +keywords:"data center") (+keywords:"Facebook" +keywords:"data center")`,
		true,
	},
	{
		// Search._id: "53a8141bfc16f8060f000007"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16' OR '2014-06-15' OR '2014-06-14' OR '2014-06-13' OR '2014-06-12' OR '2014-06-11' OR '2014-06-10' OR '2014-06-09' OR '2014-06-08' OR '2014-06-07' OR '2014-06-06' OR '2014-06-05' OR '2014-06-04' OR '2014-06-03' OR '2014-06-02' OR '2014-06-01' OR '2014-05-31' OR '2014-05-30' OR '2014-05-29' OR '2014-05-28' OR '2014-05-27' OR '2014-05-26' OR '2014-05-25' OR '2014-05-24' OR '2014-05-23') AND keywords:((("FedRAMP")))`,
		`+keywords:"FedRAMP"`,
		false,
	},
	{
		// Search._id: "53a816e8fc16f8060f00000a"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("big data")))`,
		`+keywords:"big data"`,
		true,
	},
	{
		// Search._id: "53a816e8fc16f8060f000012"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("cybersecurity")))`,
		`+keywords:"cybersecurity"`,
		false,
	},
	{
		// Search._id: "53a816e8fc16f8060f00000e"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("mobility")))`,
		`+keywords:"mobility"`,
		false,
	},
	{
		// Search._id: "53a816e8fc16f8060f00000c"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("data center")))`,
		`+keywords:"data center"`,
		true,
	},
	{
		// Search._id: "53a816e8fc16f8060f000010"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("cloud")))`,
		`+keywords:"cloud"`,
		false,
	},
	{
		// Search._id: "53a83074fc16f80ce3000001"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:((("CDW") OR ("CDW-G") OR ("CDWG")) NOT (("collision damage waiver") OR ("CDWR") OR ("California Department of Water Resources")))`,
		`keywords:"CDW" keywords:"CDW-G" keywords:"CDWG"`,
		true,
	},
	{
		// Search._id: "53a840abfc16f8196c000001"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:(("EMC"))`,
		`+keywords:"EMC"`,
		false,
	},
	{
		// Search._id: "53a84470fc16f8196c000003"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:((("CDW") OR ("CDW-G") OR ("CDWG")) NOT ("collision damage waiver"))`,
		`keywords:"CDW" keywords:"CDW-G" keywords:"CDWG"`,
		true,
	},
	{
		// Search._id: "53a8472efc16f8196c000005"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:(("FedRAMP"))`,
		`+keywords:"FedRAMP"`,
		false,
	},
	{
		// Search._id: "53a96861fc16f86be1000020"
		`published:('2014-06-24' OR '2014-06-23') AND keywords:(("cybersecurity"))`,
		`+keywords:"cybersecurity"`,
		false,
	},
	{
		// Search._id: "53a84decfc16f8196c00000b"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:(("cloud"))`,
		`+keywords:"cloud"`,
		false,
	},
	{
		// Search._id: "53a86667fc16f82001000001"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16' OR '2014-06-15' OR '2014-06-14' OR '2014-06-13' OR '2014-06-12' OR '2014-06-11' OR '2014-06-10' OR '2014-06-09' OR '2014-06-08' OR '2014-06-07' OR '2014-06-06' OR '2014-06-05' OR '2014-06-04' OR '2014-06-03' OR '2014-06-02' OR '2014-06-01' OR '2014-05-31' OR '2014-05-30' OR '2014-05-29' OR '2014-05-28' OR '2014-05-27' OR '2014-05-26' OR '2014-05-25' OR '2014-05-24' OR '2014-05-23') AND keywords:(("cloud"))`,
		`+keywords:"cloud"`,
		false,
	},
	{
		// Search._id: "53a87ffbfc16f82001000003"
		`published:('2014-06-23' OR '2014-06-22') AND keywords:((("NetApp") OR ("Red Hat") OR ("EMC") OR ("Lenovo") OR ("Lexmark") OR ("Symantec") OR ("Tripp-Lite") OR ("VMware") OR ("Cisco") OR ("Epson") OR ("HP") OR ("IBM") OR ("Citrix") OR ("Samsung") OR ("Google") OR ("3M") OR ("AirWatch") OR ("Amazon") OR ("Aruba Networks") OR ("Asus") OR ("Attachmate") OR ("Autodesk") OR ("Avaya") OR ("Aver Information Inc.") OR ("Avocent") OR ("Barracuda Networks") OR ("Belkin") OR ("Blue Coat") OR ("BMC Software") OR ("Box.com") OR ("Brocade") OR ("Brother") OR ("C2G") OR ("Canon") OR ("Check Point Software") OR ("Code Scanners") OR ("CommVault") OR ("Drobo") OR ("EDGE Tech Corp") OR ("Elo Touchsystems") OR ("Enovate IT") OR ("Ergotron") OR ("F5 Networks") OR ("Fortinet") OR ("Fujitsu") OR ("Fusion IO") OR ("Honeywell") OR ("Imation") OR ("InFocus") OR ("Intel") OR ("Isilon") OR ("Juniper Networks") OR ("Kaspersky Lab") OR ("Kensington") OR ("Kingston") OR ("Kodak Scanners") OR ("LANDesk Software") OR ("LG Electronics") OR ("Liebert") OR ("LifeSize Communications") OR ("McAfee") OR ("Meraki") OR ("Microsoft") OR ("Mitsubishi") OR ("Mobile Iron") OR ("Motion Computing") OR ("Motorola Enterprise Mobility") OR ("NComputing") OR ("NEC") OR ("NETGEAR") OR ("Nimble Storage") OR ("Nuance Communications") OR ("OKI") OR ("Optoma") OR ("Oracle Targus") OR ("Panasonic") OR ("Peerless") OR ("Planar") OR ("Plantronics") OR ("Polycom") OR ("PolyVision") OR ("Quantum") OR ("Quest Software") OR ("Raritan") OR ("Riverbed") OR ("RSA Security") OR ("Rubbermaid") OR ("Salesforce.com") OR ("SAP America") OR ("Seagate Technology") OR ("Sharp Electronics") OR ("ShoreTel") OR ("SonicWALL") OR ("Sony") OR ("Sophos") OR ("Splunk Software") OR ("StarTech.com") OR ("Trend Micro") OR ("Varonis") OR ("Veeam") OR ("ViewSonic") OR ("Vision Solutions") OR ("WatchGuard") OR ("Websense") OR ("West Point Products") OR ("Western Digital") OR ("Wyse Technology") OR ("Accenture") OR ("Best Buy") OR ("Buy.com") OR ("Rakuten") OR ("Dell") OR ("Dimension Data") OR ("ePlus") OR ("Ingram Micro") OR ("Insight Enterprises") OR ("Newegg") OR ("Office Depot") OR ("Office Max") OR ("PC Connection") OR ("PCM") OR ("Presidio") OR ("SHI") OR ("Softchoice") OR ("Staples") OR ("Tech Data") OR ("TigerDirect") OR ("World Wide Technology") OR ("&#34;Computer Associates&#34;")))`,
		`keywords:"NetApp" keywords:"Red Hat" keywords:"EMC" keywords:"Lenovo" keywords:"Lexmark" keywords:"Symantec" keywords:"Tripp-Lite" keywords:"VMware" keywords:"Cisco" keywords:"Epson" keywords:"HP" keywords:"IBM" keywords:"Citrix" keywords:"Samsung" keywords:"Google" keywords:"3M" keywords:"AirWatch" keywords:"Amazon" keywords:"Aruba Networks" keywords:"Asus" keywords:"Attachmate" keywords:"Autodesk" keywords:"Avaya" keywords:"Aver Information Inc." keywords:"Avocent" keywords:"Barracuda Networks" keywords:"Belkin" keywords:"Blue Coat" keywords:"BMC Software" keywords:"Box.com" keywords:"Brocade" keywords:"Brother" keywords:"C2G" keywords:"Canon" keywords:"Check Point Software" keywords:"Code Scanners" keywords:"CommVault" keywords:"Drobo" keywords:"EDGE Tech Corp" keywords:"Elo Touchsystems" keywords:"Enovate IT" keywords:"Ergotron" keywords:"F5 Networks" keywords:"Fortinet" keywords:"Fujitsu" keywords:"Fusion IO" keywords:"Honeywell" keywords:"Imation" keywords:"InFocus" keywords:"Intel" keywords:"Isilon" keywords:"Juniper Networks" keywords:"Kaspersky Lab" keywords:"Kensington" keywords:"Kingston" keywords:"Kodak Scanners" keywords:"LANDesk Software" keywords:"LG Electronics" keywords:"Liebert" keywords:"LifeSize Communications" keywords:"McAfee" keywords:"Meraki" keywords:"Microsoft" keywords:"Mitsubishi" keywords:"Mobile Iron" keywords:"Motion Computing" keywords:"Motorola Enterprise Mobility" keywords:"NComputing" keywords:"NEC" keywords:"NETGEAR" keywords:"Nimble Storage" keywords:"Nuance Communications" keywords:"OKI" keywords:"Optoma" keywords:"Oracle Targus" keywords:"Panasonic" keywords:"Peerless" keywords:"Planar" keywords:"Plantronics" keywords:"Polycom" keywords:"PolyVision" keywords:"Quantum" keywords:"Quest Software" keywords:"Raritan" keywords:"Riverbed" keywords:"RSA Security" keywords:"Rubbermaid" keywords:"Salesforce.com" keywords:"SAP America" keywords:"Seagate Technology" keywords:"Sharp Electronics" keywords:"ShoreTel" keywords:"SonicWALL" keywords:"Sony" keywords:"Sophos" keywords:"Splunk Software" keywords:"StarTech.com" keywords:"Trend Micro" keywords:"Varonis" keywords:"Veeam" keywords:"ViewSonic" keywords:"Vision Solutions" keywords:"WatchGuard" keywords:"Websense" keywords:"West Point Products" keywords:"Western Digital" keywords:"Wyse Technology" keywords:"Accenture" keywords:"Best Buy" keywords:"Buy.com" keywords:"Rakuten" keywords:"Dell" keywords:"Dimension Data" keywords:"ePlus" keywords:"Ingram Micro" keywords:"Insight Enterprises" keywords:"Newegg" keywords:"Office Depot" keywords:"Office Max" keywords:"PC Connection" keywords:"PCM" keywords:"Presidio" keywords:"SHI" keywords:"Softchoice" keywords:"Staples" keywords:"Tech Data" keywords:"TigerDirect" keywords:"World Wide Technology" keywords:"&#34;Computer Associates&#34;"`,
		true,
	},
	{
		// Search._id: "53a8a399fc16f86be1000001"
		`published:('2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16') AND keywords:(("EMC")) AND publicationid:(5272624d800b8e3d940006ed OR 52726259800b8e3d94000bf9 OR 5294f834a75ae411660059b4 OR 52726255800b8e3d94000a4f OR 5272625684e7533dd00002e3 OR 52726259800b8e3d94000bbb OR 5272625aa75ae43e05000191 OR 5272625ba75ae43e050001a7 OR 5272625884e7533dd0000323 OR 52c74079800b8e3387005eb6 OR 52726256800b8e3d94000a8f OR 5294d213800b8e10670010a9 OR 52a704d5800b8e67a0014867 OR 52a7045a84e753686f0151d4 OR 52726256800b8e3d94000aab OR 52726256800b8e3d94000a87 OR 5272625ba75ae43e0500019d OR 52726255800b8e3d94000a39 OR 5272625584e7533dd00002bf OR 52726256800b8e3d94000a97 OR 52726256800b8e3d94000a67 OR 52726257800b8e3d94000af3 OR 52726254800b8e3d94000a05 OR 5272625884e7533dd0000327 OR 529501eeb6bbac2346003694 OR 5294faff84e75310ef00335e OR 5272624684e7533dd00000cf OR 52726256b6bbac3d58000147 OR 527262574113de3d8a0000fb OR 527262494113de3d8a000051 OR 52c73eec84e7533371008784 OR 52851cc872cd3f15e400eb26 OR 535669644113de7639000001)`,
		`+keywords:"EMC"`,
		false,
	},
	{
		// Search._id: "53a92c32fc16f86be1000003"
		`published:('2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17') AND keywords:(("citrix")) AND publicationid:(5272625584e7533dd00002c3 OR 536aacc04113de7749000001 OR 5272625984e7533dd000033f OR 5294f7f5a75ae4116600598e OR 52726259800b8e3d94000bf9 OR 5294f834a75ae411660059b4 OR 52726255800b8e3d94000a4f OR 5272625684e7533dd00002e3 OR 52726259800b8e3d94000bbb OR 5272625884e7533dd0000323 OR 52c74079800b8e3387005eb6 OR 5272625aa75ae43e05000191 OR 52726256800b8e3d94000a8f OR 52a7045a84e753686f0151d4 OR 52726256800b8e3d94000a87 OR 52726256800b8e3d94000aab OR 5272625584e7533dd00002bf OR 52726254800b8e3d94000a05 OR 5272625884e7533dd0000327 OR 529501eeb6bbac2346003694 OR 535669644113de7639000001 OR 52726254800b8e3d94000a0d OR 52726254800b8e3d940009fd OR 5272624684e7533dd00000cf OR 52726256b6bbac3d58000147 OR 52726256800b8e3d94000a97 OR 52950479b6bbac23460037f2 OR 529507694113de10cd003d89 OR 529509eaa75ae41166007516 OR 52950ca7a75ae41166007d1f OR 52726256800b8e3d94000a9b OR 52950dc4a75ae41166007f56)`,
		`+keywords:"citrix"`,
		false,
	},
	{
		// Search._id: "53a92fb6fc16f86be1000005"
		`published:('2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17') AND keywords:(("Oracle")) AND publicationid:(52726242800b8e3d9400033f OR 5294d611a75ae411660027f3 OR 5272624a72cd3f3d71000031 OR 5294d6d584e75310ef00171e OR 5294d92472cd3f10790022ca OR 5294d97fa75ae41166002d48 OR 5294db15b6bbac23460019e9 OR 5294db644113de10cd001bc6 OR 5294dbc34113de10cd001c05 OR 5294dacca75ae41166002f03 OR 5294dc7ab6bbac2346001a8e OR 5294dcfa72cd3f1079002591 OR 52726243800b8e3d9400036f OR 52726259800b8e3d94000bfd OR 5294dda4800b8e1067001653 OR 5294de5184e75310ef001ca5 OR 5339b57c800b8e6094000001 OR 5294df3972cd3f1079002815 OR 52726251800b8e3d94000955 OR 5294e3e4b6bbac2346001f93 OR 5294e4cd84e75310ef002247 OR 5272624e4113de3d8a0000b9)`,
		`+keywords:"Oracle"`,
		false,
	},
	{
		// Search._id: "53a93339fc16f86be1000007"
		`published:('2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17') AND keywords:(("Oracle")) AND publicationid:(5294e4f6a75ae41166003d88 OR 5272624b4113de3d8a00008f OR 528e615c84e7536d52003763 OR 5272624a4113de3d8a000059 OR 52726254800b8e3d94000a01 OR 5339e21f2338dc693a000001 OR 527262564113de3d8a0000e7 OR 52851cc872cd3f15e400eb26 OR 527262564113de3d8a0000f7 OR 5272625384e7533dd0000237 OR 5272625b800b8e3d94000c1d OR 52726253800b8e3d940009c1 OR 52726254800b8e3d940009dd OR 5294ef8fa75ae4116600511c OR 5272624bb6bbac3d580000fb OR 5339b57c800b8e6094000001 OR 52726247800b8e3d940004eb OR 5272624f84e7533dd000018b OR 5272625b800b8e3d94000c25)`,
		`+keywords:"Oracle"`,
		false,
	},
	{
		// Search._id: "53a936bdfc16f86be1000009"
		`published:('2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17') AND keywords:(("Oracle")) AND publicationid:(52726248800b8e3d94000503 OR 5272624f800b8e3d940007f7 OR 5272625772cd3f3d710000cb OR 5272624e84e7533dd0000177 OR 52726255b6bbac3d5800013b OR 52726256800b8e3d94000aaf OR 527262524113de3d8a0000db)`,
		`+keywords:"Oracle"`,
		false,
	},
	{
		`published:('2014-06-27' OR '2014-06-26' OR '2014-06-25' OR '2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16' OR '2014-06-15' OR '2014-06-14' OR '2014-06-13' OR '2014-06-12' OR '2014-06-11' OR '2014-06-10' OR '2014-06-09' OR '2014-06-08' OR '2014-06-07' OR '2014-06-06' OR '2014-06-05' OR '2014-06-04' OR '2014-06-03' OR '2014-06-02' OR '2014-06-01' OR '2014-05-31' OR '2014-05-30' OR '2014-05-29' OR '2014-05-28') AND keywords:(("amazon" AND "cloud computing"))`,
		`+keywords:"amazon" +keywords:"cloud computing"`,
		true,
	},
	{
		`published:('2014-06-30' OR '2014-06-29' OR '2014-06-28' OR '2014-06-27' OR '2014-06-26' OR '2014-06-25' OR '2014-06-24' OR '2014-06-23' OR '2014-06-22' OR '2014-06-21' OR '2014-06-20' OR '2014-06-19' OR '2014-06-18' OR '2014-06-17' OR '2014-06-16' OR '2014-06-15' OR '2014-06-14' OR '2014-06-13' OR '2014-06-12' OR '2014-06-11' OR '2014-06-10' OR '2014-06-09' OR '2014-06-08' OR '2014-06-07' OR '2014-06-06' OR '2014-06-05' OR '2014-06-04' OR '2014-06-03' OR '2014-06-02' OR '2014-06-01' OR '2014-05-31') AND keywords:(("Apple"))`,
		`+keywords:"Apple"`,
		false,
	},
}
