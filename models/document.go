package models

import "encoding/xml"

type Documents struct {
	XMLName   xml.Name   `xml:"ProQuestExport"`
	Documents []Document `xml:"Documents>Literature"`
}

type Document struct {
	// XMLName xml.Name `xml:"DocInfo"`
	Type                 string           `json:"_type"`
	TitleInfo            []TitleInfo      `xml:"DocInfo>TitleInfo>Title"`
	AlternateTitle       []TitleInfo      `xml:"DocInfo>TitleInfo>AlternateTitle"`
	AccessionNumber      int              `xml:"DocInfo>AccessionNumber"`
	DatabaseName         string           `xml:"DocInfo>DatabaseName"`
	ProquestID           int              `xml:"DocInfo>ProquestID"`
	DocumentIDs          []DocumentID     `xml:"DocInfo>DocumentIDs>DocumentID"`
	SourceAttribution    string           `xml:"DocInfo>SourceAttribution"`
	PublicationDate      string           `xml:"DocInfo>PublicationDate"`
	PublicationAlphaDate string           `xml:"DocInfo>PublicationAlphaDate"`
	DateCreated          string           `xml:"DocInfo>DateCreated"`
	DateRevised          string           `xml:"DocInfo>DateRevised"`
	FirstAvailableDate   string           `xml:"DocInfo>FirstAvailableDate"`
	LastUpdateDate       string           `xml:"DocInfo>LastUpdateDate"`
	DocumentStatus       string           `xml:"DocInfo>DocumentStatus"`
	DocumentType         string           `xml:"DocInfo>DocumentType"`
	SourceType           string           `xml:"DocInfo>SourceType"`
	DOI                  string           `xml:"DocInfo>DOI,omitempty"` // omitempty?
	Language             string           `xml:"DocInfo>Language"`
	NumRefs              int              `xml:"DocInfo>DocFeatures>NumRefs"`
	URL                  string           `xml:"DocInfo>URL"`
	Contributors         []Contributor    `xml:"Contributors>Contributor"`
	Abstract             []Abstract       `xml:"Abstract"`
	HeadingTerms         []HeadingTerm    `xml:"Subjects>HeadingTerms>HeadingTerm"`
	SubjectTerms         []string         `xml:"Subjects>SubjectTerms>SubjectTerm"`
	IdentifierTerms      []string         `xml:"Subjects>IdentifierTerms>IdentifierTerm"`
	Classifications      []Classification `xml:"Classifications>ClassTerm"`
	SubstanceTerms       []SubstanceTerm  `xml:"SubstanceInfo>SubstanceTerms>SubstanceTerm"`
	PublicationInfo      PublicationInfo  `xml:"PublicationInfo"`
}
type TitleInfo struct {
	Language string `xml:",attr"`
	Title    string `xml:",chardata"`
}
type DocumentID struct {
	Type string `xml:"IDType,attr"`
	ID   string `xml:",chardata"`
}
type Contributor struct {
	Order          int      `xml:"ContribOrder,attr"`
	Role           string   `xml:"ContribRole,attr"`
	NormalizedName string   `xml:"NormalizedName"`
	LastName       string   `xml:"LastName"`
	FirstName      string   `xml:"FirstName"`
	CompanyName    []string `xml:"ContribCompanyName"`
	EmailAddress   string   `xml:"EmailAddress,omitempty"`
	RefCode        RefCode  `xml:"RefCode,omitempty"`
	PersonTitle    string   `xml:"PersonTitle,omitempty"`
	NameSuffix     string   `xml:"NameSuffix,omitempty"`
}
type RefCode struct {
	Type string `xml:"RefCodeType,attr"`
	ID   string `xml:",chardata"`
}
type Abstract struct {
	Text      string `xml:",chardata"`
	WordCount string `xml:"WordCount,attr"`
	Type      string `xml:"AbstractType,attr"`
	Language  string `xml:"Language,attr"`
}
type HeadingTerm struct {
	TermVocab            string           `xml:"TermVocab,attr"`
	HeadingTermType      string           `xml:"HeadingTermType,attr"`
	Heading              Heading          `xml:"Heading"`
	HeadingQualifier     HeadingQualifier `xml:"HeadingQualifier"`
	QualifierNameSubLink string           `xml:"QualifierNameSubLink"`
}
type Heading struct {
	Text        string `xml:",chardata"`
	MajorTopic  string `xml:"MajorTopic,attr"`
	HeadingType string `xml:"HeadingType,attr"`
}
type HeadingQualifier struct {
	Text                 string `xml:",chardata"`
	HeadingQualifierType string `xml:"HeadingQualifierType,attr"`
}
type Classification struct {
	TermVocab      string `xml:"TermVocab,attr"`
	ClassTermType  string `xml:"ClassTermType,attr"`
	ClassCode      int    `xml:"ClassCode"`
	ClassExpansion string `xml:"ClassExpansion"`
}
type SubstanceTerm struct {
	SubstanceName    string            `xml:"SubstanceName"`
	SubstanceNumbers []SubstanceNumber `xml:"SubstanceNumber"`
}
type SubstanceNumber struct {
	Text string `xml:",chardata"`
	Type string `xml:"SubstanceNumberType,attr"`
}
type PublicationInfo struct {
	Title      string    `xml:"PublicationTitle"`
	Volume     string    `xml:"Volume"`
	Issue      string    `xml:"Issue"`
	Locators   []Locator `xml:"Locators>LocatorID"`
	IssueTitle string    `xml:"IssueTitle,omitempty"`
	Publisher  Publisher `xml:"Publisher"`
	Pages      []Page    `xml:"Pages"`
	Notes      []Note    `xml:"PublicationNotes>PublicationNote"`
}
type Locator struct {
	Text string `xml:",chardata"`
	Type string `xml:"IDType,attr"`
}
type Publisher struct {
	Name     string            `xml:"PublisherName"`
	Location PublisherLocation `xml:"PublisherLocation"`
}
type PublisherLocation struct {
	MailingAddress string `xml:"PublisherMailingAddress"`
	EmailAddress   string `xml:"PublisherEmailAddress"`
	CityName       string `xml:"PublisherCityName"`
	PostCode       string `xml:"PublisherPostCode"`
	CountryName    string `xml:"PublisherCountryName"`
}
type Page struct {
	StartPage  string `xml:"StartPage"`
	EndPage    string `xml:"EndPage"`
	Pagination string `xml:"Pagination"`
}
type Note struct {
	Text string `xml:",chardata"`
	Type string `xml:"NoteType,attr"`
}
