package models

import (
	"encoding/xml"
	"time"
)

type Documents struct {
	XMLName   xml.Name   `xml:"ProQuestExport"`
	AlertID   string     `xml:"Options>AlertID"`
	AlertName string     `xml:"Options>AlertName"`
	Timestamp string     `xml:"Options>Timestamp" json:"-"` // for date format: https://stackoverflow.com/questions/17301149/golang-xml-unmarshal-and-time-time-fields
	Documents []Document `xml:"Documents>Literature"`
}
type Document struct {
	// Type                 string           `json:"type"` // once Index is embase, there is no need for an extra type
	IndexedAt            time.Time        `json:"indexedAt"`
	AlertID              string           `json:"alertID"`
	AlertName            string           `json:"alertName"`
	TitleInfo            []TitleInfo      `xml:"DocInfo>TitleInfo>Title"`
	AlternateTitle       []TitleInfo      `xml:"DocInfo>TitleInfo>AlternateTitle" json:",omitempty"`
	AccessionNumber      string           `xml:"DocInfo>AccessionNumber"`
	DatabaseName         string           `xml:"DocInfo>DatabaseName"`
	ProquestID           int              `xml:"DocInfo>ProquestID"`
	DocumentIDs          []DocumentID     `xml:"DocInfo>DocumentIDs>DocumentID" json:",omitempty"`
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
	DOI                  string           `xml:"DocInfo>DOI,omitempty" json:",omitempty"`
	Language             string           `xml:"DocInfo>Language"`
	NumRefs              int              `xml:"DocInfo>DocFeatures>NumRefs" json:",omitempty"`
	URL                  string           `xml:"DocInfo>URL"`
	OutboundLinks        []string         `xml:"DocInfo>OutboundLinks" json:",omitempty"`
	Contributors         []Contributor    `xml:"Contributors>Contributor" json:",omitempty"`
	Abstract             []Abstract       `xml:"Abstract"`
	HeadingTerms         []HeadingTerm    `xml:"Subjects>HeadingTerms>HeadingTerm" json:",omitempty"`
	SubjectTerms         []string         `xml:"Subjects>SubjectTerms>SubjectTerm" json:",omitempty"`
	IdentifierTerms      []string         `xml:"Subjects>IdentifierTerms>IdentifierTerm" json:",omitempty"`
	Classifications      []Classification `xml:"Classifications>ClassTerm" json:",omitempty"`
	ConferenceInfo       ConferenceInfo   `xml:"ConferenceInfo" json:",omitempty"`
	SubstanceTerms       []SubstanceTerm  `xml:"SubstanceInfo>SubstanceTerms>SubstanceTerm" json:",omitempty"`
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
	LastName       string   `xml:"LastName" json:",omitempty"`
	FirstName      string   `xml:"FirstName" json:",omitempty"`
	CompanyName    []string `xml:"ContribCompanyName" json:",omitempty"`
	EmailAddress   string   `xml:"EmailAddress" json:",omitempty"`
	RefCode        RefCode  `xml:"RefCode" json:",omitempty"`
	PersonTitle    string   `xml:"PersonTitle" json:",omitempty"`
	NameSuffix     string   `xml:"NameSuffix" json:",omitempty"`
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
	HeadingQualifier     HeadingQualifier `xml:"HeadingQualifier" json:",omitempty"`
	QualifierNameSubLink string           `xml:"QualifierNameSubLink" json:",omitempty"`
}
type Heading struct {
	Text        string `xml:",chardata"`
	MajorTopic  string `xml:",attr"`
	HeadingType string `xml:",attr"`
}
type HeadingQualifier struct {
	Text                 string `xml:",chardata"`
	HeadingQualifierType string `xml:",attr"`
}
type Classification struct {
	TermVocab      string `xml:"TermVocab,attr"`
	ClassTermType  string `xml:"ClassTermType,attr"`
	ClassCode      int    `xml:"ClassCode"`
	ClassExpansion string `xml:"ClassExpansion"`
}
type ConferenceInfo struct {
	Title    string           `xml:"ConferenceTitle"`
	Country  string           `xml:"ConferenceLocationInfo>ConferenceCountry"`
	Location string           `xml:"ConferenceLocationInfo>ConferenceLocation"`
	Dates    []ConferenceDate `xml:"ConferenceDates>ConferenceDate"`
}
type ConferenceDate struct {
	Type string `xml:"ConferenceDateType,attr"`
	Date string `xml:",chardata"`
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
	Supplement string    `xml:"Supplement" json:",omitempty"`
	Issue      string    `xml:"Issue"`
	Locators   []Locator `xml:"Locators>LocatorID"`
	IssueTitle string    `xml:"IssueTitle" json:",omitempty"`
	Publisher  Publisher `xml:"Publisher" json:",omitempty"`
	Pages      []Page    `xml:"Pages"`
	Notes      []Note    `xml:"PublicationNotes>PublicationNote"`
}
type Locator struct {
	Text string `xml:",chardata"`
	Type string `xml:"IDType,attr"`
}
type Publisher struct {
	Name     string            `xml:"PublisherName" json:",omitempty"`
	Location PublisherLocation `xml:"PublisherLocation" json:",omitempty"`
}
type PublisherLocation struct {
	MailingAddress string `xml:"PublisherMailingAddress" json:",omitempty"`
	EmailAddress   string `xml:"PublisherEmailAddress" json:",omitempty"`
	CityName       string `xml:"PublisherCityName" json:",omitempty"`
	PostCode       string `xml:"PublisherPostCode" json:",omitempty"`
	CountryName    string `xml:"PublisherCountryName" json:",omitempty"`
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
