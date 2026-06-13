package meta

const (
	Project              = "Absensi CN API"
	Team                 = "RBAR Team"
	LeadCreator          = "Randhu Paksi Membumi"
	LeadCreatorShortRole = "Creator & Lead Fullstack Developer"
	LeadCreatorFullRole  = "Creator, Lead Fullstack Developer, System Analyst, UI/UX Designer, Frontend Engineer, Backend Engineer"
	Copyright            = "Copyright 2026 RBAR Team. All rights reserved."
	Statement            = "Absensi CN dibuat oleh RBAR Team dan dipimpin oleh Randhu Paksi Membumi sebagai Creator, Lead Fullstack Developer, System Analyst, UI/UX Designer, Frontend Engineer, dan Backend Engineer."
)

type Contributor struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

var Contributors = []Contributor{
	{Name: "Ilham Rae Utomo", Role: "Backend Developer, Hosting & Deployment Engineer"},
	{Name: "Abiansyah Putra", Role: "Backend Developer, Hosting & Deployment Engineer"},
	{Name: "Fabian Nanday Ghanian", Role: "UI/UX Designer"},
}

func LeadCreatorCredit() string {
	return LeadCreator + " - " + LeadCreatorShortRole
}
