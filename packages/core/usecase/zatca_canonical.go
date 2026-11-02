package usecase

import (
	"fmt"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	zatcaCAC = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
	zatcaCBC = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
	zatcaEXT = "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"
)

// CanonicalizeXML applies the three ZATCA invoice reference filters followed by
// inclusive C14N 1.1 without comments. Prefixes and document whitespace survive.
func CanonicalizeXML(xmlData []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlData); err != nil {
		return nil, fmt.Errorf("invalid invoice XML: %w", err)
	}
	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("invoice XML has no root element")
	}
	filterZatcaNodes(root)
	return dsig.MakeC14N11Canonicalizer().Canonicalize(root)
}

func filterZatcaNodes(parent *etree.Element) {
	for _, child := range parent.ChildElements() {
		ns := child.NamespaceURI()
		excluded := (ns == zatcaEXT && child.Tag == "UBLExtensions") || (ns == zatcaCAC && child.Tag == "Signature")
		if ns == zatcaCAC && child.Tag == "AdditionalDocumentReference" {
			for _, id := range child.ChildElements() {
				if id.NamespaceURI() == zatcaCBC && id.Tag == "ID" && id.Text() == "QR" {
					excluded = true
				}
			}
		}
		if excluded {
			parent.RemoveChild(child)
		} else {
			filterZatcaNodes(child)
		}
	}
}
