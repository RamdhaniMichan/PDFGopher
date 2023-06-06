package pdfgopher

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"reflect"

	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/image/draw"
)

// FileType represents the type of file.
type FileType string

// Option is a function type used for applying options to a PDFProcessor.
// It takes a pointer to a PDFProcessor and modifies it according to the provided option.
type Option func(*PDFProcessor)

// Constants for different file types.
const (
	PDF      FileType = "pdf"
	Image    FileType = "image"
	Document FileType = "document"
)

// PDFProcessor provides operations related to PDF files.
type PDFProcessor struct {
	FilePath      string
	Base64Output  string
	PDFProtection bool
	*OptionFilePDF
	*OptionMetadataPDF
}

// OptionMetadataPDF represents options for modifying PDF metadata.
type OptionMetadataPDF struct {
	Title   string
	Author  string
	Subject string
}

// OptionFilePDF represents options for working with PDF files.
type OptionFilePDF struct {
	PasswordPDF   string
	QRCodePath    string
	StampPosition string
}

// NewPDFGopher constructor to retrieve struct PDFProcessor
func NewPDFGopher(filePath string, options ...Option) *PDFProcessor {
	option := &PDFProcessor{
		FilePath: filePath,
		OptionFilePDF: &OptionFilePDF{
			StampPosition: "br",
		},
		OptionMetadataPDF: &OptionMetadataPDF{},
	}

	for _, opt := range options {
		opt(option)
	}

	return option
}

// WithOptionMetadataPDF returns an Option function that sets the OptionMetadataPDF value.
func WithOptionMetadataPDF(value OptionMetadataPDF) Option {
	return func(p *PDFProcessor) {
		p.OptionMetadataPDF = &value
	}
}

// WithOptionFilePDF returns an Option function that sets the OptionFilePDF value.
func WithOptionFilePDF(value OptionFilePDF) Option {
	return func(p *PDFProcessor) {
		v := reflect.ValueOf(value)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.String() != "" {
				reflect.ValueOf(p.OptionFilePDF).Elem().Field(i).Set(field)
			}
		}
	}
}

// ProcessFile processes the input file based on its type.
func (p *PDFProcessor) ProcessFile() error {
	fileType := getFileType(p.FilePath)
	switch fileType {
	case PDF:
		// Check if the PDF file has a password
		hasPassword, err := hasPDFPassword(p.FilePath, p.PasswordPDF)
		if err != nil {
			return err
		}

		p.PDFProtection = hasPassword

		if hasPassword {
			// Descrypt the PDF File
			err := decrypted(p.FilePath, p.PasswordPDF)
			if err != nil {
				return err
			}
		}

		// Process the PDF file
		err = p.processPDF(p.FilePath, p.OptionFilePDF.QRCodePath, p.OptionFilePDF.StampPosition)
		if err != nil {
			return err
		}

		// // Delete the temporary PDF file
		// err = os.Remove(filePath)
		// if err != nil {
		// 	return "", err
		// }
	case Image:
		// Convert the image file to PDF
		pdfFilePath, err := convertImageToPDF(p.FilePath)
		if err != nil {
			return err
		}

		// Process the converted PDF file
		err = p.processPDF(pdfFilePath, p.OptionFilePDF.QRCodePath, p.OptionFilePDF.StampPosition)
		if err != nil {
			return err
		}

		// // Delete the temporary PDF file
		// err = os.Remove(pdfFilePath)
		// if err != nil {
		// 	return "", err
		// }
	case Document:
		// Convert the document file to PDF
		pdfFilePath, err := convertDocumentToPDF(p.FilePath)
		if err != nil {
			return err
		}

		// Process the converted PDF file
		err = p.processPDF(pdfFilePath, p.OptionFilePDF.QRCodePath, p.OptionFilePDF.StampPosition)
		if err != nil {
			return err
		}

		// Delete the temporary PDF file
		err = os.Remove(pdfFilePath)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}

// pdfToBase64 converts a PDF file to base64 encoding.
func (p *PDFProcessor) pdfToBase64(filePath string) error {
	// Read the PDF file into memory.
	pdfFile, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	// Encode the PDF file as base64.
	base64Output := base64.StdEncoding.EncodeToString(pdfFile)
	// Set the Base64Output field of the PDFProcessor struct.
	p.Base64Output = base64Output

	return nil
}

// getFileType returns the type of file based on its extension.
func getFileType(filePath string) FileType {
	extension := strings.ToLower(filepath.Ext(filePath))
	switch extension {
	case ".pdf":
		return PDF
	case ".jpg", ".jpeg", ".png":
		return Image
	case ".doc", ".docx":
		return Document
	default:
		return ""
	}
}

// hasPDFPassword checks if the PDF file is password-protected.
func hasPDFPassword(filePath string, password string) (bool, error) {
	command := ""
	if password != "" {
		command = fmt.Sprintf("pdfcpu validate -mode=quiet -upw='%s' %s", password, filePath)
	} else {
		command = fmt.Sprintf("pdfcpu validate %s", filePath)
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 1 {
			// PDF is password protected
			return true, nil
		} else {
			// Other execution error
			return false, exitError
		}
	} else {
		// PDF is not password protected
		return false, err
	}
}

// decrypted unction is used to remove the protection from a PDF file by decrypting it with a provided password.
func decrypted(filePath string, password string) error {
	command := fmt.Sprintf("pdfcpu decrypt -upw %s %s", password, filePath)

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing pdfcpu command: %s\n", err.Error())
		return err
	}

	return nil

}

// encrypted function is used to encrypt a previously decrypted PDF.
func encrypted(filePath string, password string) error {
	command := fmt.Sprintf("pdfcpu encrypt -upw %s -opw %s %s", password, password, filePath)

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing pdfcpu command: %s\n", err.Error())
		return err
	}

	return nil
}

// processPDF performs operations on the PDF file using pdfcpu-cli.
func (p *PDFProcessor) processPDF(filePath string, qrCode string, stampPosition string) error {
	// Add QR code to the PDF file
	err := addQRCodeToPDF(filePath, qrCode, stampPosition)
	if err != nil {
		return err
	}

	//add metadata to file pdf
	if !IsStructEmpty(p.OptionMetadataPDF) {
		err := addedMetadata(filePath, p.OptionMetadataPDF)
		if err != nil {
			return err
		}
	}

	//add protection to file pdf
	if p.PDFProtection {
		err := encrypted(filePath, p.OptionFilePDF.PasswordPDF)
		if err != nil {
			return err
		}
	}

	//Convert pdf file to base64 as output file
	err = p.pdfToBase64(filePath)
	if err != nil {
		return err
	}

	return nil
}

// addedMetadata to add metadata into a pdf file.
func addedMetadata(filePath string, metadata *OptionMetadataPDF) error {
	command := fmt.Sprintf("pdfcpu properties add %s 'Title = %s' 'Author = %s' 'Subject = %s'", filePath, metadata.Title, metadata.Author, metadata.Subject)

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// addQRCodeToPDF adds a QR code to the PDF file using pdfcpu-cli.
func addQRCodeToPDF(filePath string, qrCode string, stampPosition string) error {
	if qrCode == "" {
		return errors.New("QR Code is empty")
	}

	// Load the icon image
	iconFile, err := os.Open(qrCode)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		} else {
			return fmt.Errorf("error opening file: %s", err.Error())
		}
	}

	defer iconFile.Close()

	command := fmt.Sprintf("pdfcpu stamp add -pages even,odd -mode image -- '%s' 'pos:%s, rot:0, sc:.1' %s", iconFile.Name(), stampPosition, filePath)

	// Execute the command
	cmd := exec.Command("sh", "-c", command)

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Command exited with a non-zero status
			fmt.Printf("Command failed with error: %s\n", exitError.Error())
			if len(exitError.Stderr) > 0 {
				return fmt.Errorf("error output:%s", string(exitError.Stderr))
			}
		} else {
			// Other execution error
			return fmt.Errorf("error executing pdfcpu command: %s", err.Error())
		}
	}

	return nil
}

// generateThumbnailFromPDF generates a thumbnail from the PDF.
// func generateThumbnailFromPDF(filePath string) error {
// 	panic("implement me")
// }

// GenerateQRCodeWithIcon generate QR Code with icon in the center position.
func GenerateQRCodeWithIcon(data string, iconPath string, filePath string) (string, error) {
	// Create a new QR code barcode with the given data
	qrCode, err := qr.Encode(data, qr.M, qr.Auto)
	if err != nil {
		return "", err
	}

	// Scale the barcode to the desired size
	qrCode, err = barcode.Scale(qrCode, 125, 125)
	if err != nil {
		return "", err
	}

	// Load the icon image
	iconFile, err := os.Open(iconPath)
	if err != nil {
		return "", err
	}
	defer iconFile.Close()

	iconImg, _, err := image.Decode(iconFile)
	if err != nil {
		return "", err
	}

	resizeIcon := image.NewRGBA(image.Rect(0, 0, 30, 30))

	draw.CatmullRom.Scale(resizeIcon, resizeIcon.Bounds(), iconImg, iconImg.Bounds(), draw.Over, nil)

	// Create a new image with transparent background
	finalImg := image.NewRGBA(qrCode.Bounds())

	// Calculate the position to place the icon in the center of the QR code
	iconX := (qrCode.Bounds().Max.X - resizeIcon.Bounds().Max.X) / 2
	iconY := (qrCode.Bounds().Max.Y - resizeIcon.Bounds().Max.Y) / 2

	// Draw the QR code onto the final image
	draw.Draw(finalImg, qrCode.Bounds().Add(image.Point{}), qrCode, image.Point{}, draw.Over)

	// Draw the icon onto the final image
	draw.Draw(finalImg, resizeIcon.Bounds().Add(image.Pt(iconX, iconY)), resizeIcon, image.Point{}, draw.Over)

	// Create a new file to save the QR code image with the icon
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Save the final image as a PNG file
	err = png.Encode(file, finalImg)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// convertImageToPDF converts an image file to PDF using package gofpdf.
func convertImageToPDF(imageFilePath string) (string, error) {
	// Open the input image file
	outputFile := fmt.Sprintf("process-%s", changeFileExtension(imageFilePath, "pdf"))
	file, err := os.Open(imageFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the image file
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a new page
	pdf.AddPage()

	// Calculate the aspect ratio of the image
	aspectRatio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())

	// Set the image size to fit the page width
	pageWidth, pageHeight := pdf.GetPageSize()
	imageWidth := pageWidth
	imageHeight := imageWidth / aspectRatio

	// Calculate the vertical position to center the image
	imageY := (pageHeight - imageHeight) / 2

	// Add the image to the PDF
	pdf.ImageOptions(imageFilePath, 0, imageY, imageWidth, imageHeight, false, gofpdf.ImageOptions{}, 0, "")

	// Save the PDF to the output file
	err = pdf.OutputFileAndClose(outputFile)
	if err != nil {
		return "", err
	}

	return outputFile, nil
}

// convertDocumentToPDF converts a document file to PDF using pdfcpu-cli.
func convertDocumentToPDF(documentFilePath string) (string, error) {
	panic("implement me")
}

// changeFileExtension changes the file extension to the new extension.
func changeFileExtension(filePath string, newExtension string) string {
	fileName := filepath.Base(filePath)
	fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	newFileName := fileNameWithoutExt + "." + newExtension
	return filepath.Join(filepath.Dir(filePath), newFileName)
}

// IsStructEmpty a function to check if a struct is empty or not.
func IsStructEmpty(data interface{}) bool {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.String() != "" {
			return false
		}
	}

	return true
}
