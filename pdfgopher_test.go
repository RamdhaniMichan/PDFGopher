package pdfgopher_test

import (
	"fmt"
	"testing"

	. "github.com/rdhn/PDFGopher"

	"github.com/stretchr/testify/assert"
)

func TestProcessPDF(t *testing.T) {
	pdfProcess := NewPDFGopher("./sample_pdf/process-tree-736885__480.pdf",
		WithOptionMetadataPDF(OptionMetadataPDF{Title: "Me to", Author: "Me to", Subject: "Me to"}),
		WithOptionFilePDF(OptionFilePDF{QRCodePath: "./sample_image/qr-generate.png", StampPosition: "tl"}),
	)

	// Process file
	err := pdfProcess.ProcessFile()

	fmt.Printf("pdfProcess.Base64Output: %v\n", pdfProcess.Base64Output)

	assert.NoError(t, err)
	assert.NotEmpty(t, pdfProcess.Base64Output)

}

func TestGenerateQRCode(t *testing.T) {
	qrCode, err := GenerateQRCodeWithIcon("https://google.com", "./sample_image/privyid-favicon.png", "./sample_image/qr-generate.png")

	assert.NoError(t, err)
	assert.NotEmpty(t, qrCode)
}
