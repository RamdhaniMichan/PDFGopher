
# PDFGopher Library Documentation

The PDFGopher library provides operations for processing PDF files, including adding QR codes, applying metadata, and encrypting or decrypting PDF files. It utilizes the pdfcpu-cli package for executing PDF-related commands.




## Installation

To use the PDFGopher library, you need to have Go installed. You can install the library using the following command:
```bash
 go get github.com/RamdhaniMichan/PDFGopher
```
    
## Usage
The PDFGopher library offers the following features:

### 1. Creating a PDFProcessor Instance
To begin using the library, create an instance of the PDFProcessor struct by calling the NewPDFGopher function. Pass the file path of the PDF file you want to process as the first argument. You can also provide additional options using variadic arguments.

Example:

```bash
processor := NewPDFGopher("path/to/file.pdf",
WithOptionMetadataPDF(OptionMetadataPDF{
    Title:   "Sample Title",
    Author:  "John Doe",
    Subject: "Sample Subject"}),
WithOptionFilePDF(OptionFilePDF{
    PasswordPDF:   "password123",
    QRCodePath:    "path/to/qrcode.png",
    StampPosition: "br"}),
)
```

### 2. Processing the PDF File
After creating the PDFProcessor instance, you can process the PDF file using the ProcessFile method. It will automatically detect the file type and perform the necessary operations.

Example:

```bash
err := processor.ProcessFile()
if err != nil {
    fmt.Println("Error processing the PDF file:", err)
}
```

### 3. Retrieving the Base64 Output
You can obtain the processed PDF file as a Base64-encoded string using the Base64Output field of the PDFProcessor struct.

Example:

```bash
base64Output := processor.Base64Output
fmt.Println("Base64 output:", base64Output)
```

### 4. Generating a QR Code
To generate a QR code and add it to the PDF, you can use the GenerateQRCodeWithIcon function. Pass the data to be encrypted into the QR code and the path to save the generated QR code image.

Example:

```bash
data := "https://www.example.com"
icon := "path/your-icon.png"
outputPath := "path/to/qrcode.png"

err := GenerateQRCodeWithIcon(data, icon, outputPath)
if err != nil {
    fmt.Println("Error generating the QR code:", err)
}

fmt.Printf("QR Code generated and saved to %s\n", outputPath)
```

Make sure to replace the values of data, icon and outputPath according to your needs. Upon execution, a QR code will be generated with the specified data and saved to the file specified by outputPath.

### 5. Customizing Options
The NewPDFGopher function allows you to provide optional metadata and file options when creating the PDFProcessor instance. Use the WithOptionMetadataPDF and WithOptionFilePDF functions to customize these options.

Example:

```bash
processor := NewPDFGopher("path/to/file.pdf",
WithOptionMetadataPDF(OptionMetadataPDF{
    Title:   "Sample Title",
    Author:  "John Doe",
    Subject: "Sample Subject"}),
WithOptionFilePDF(OptionFilePDF{
    PasswordPDF:   "password123",
    QRCodePath:    "path/to/qrcode.png",
    StampPosition: "br"}),
)
```


## File Type
The library supports the following file types:

* PDF: PDF files with or without password protection.
* Image: Common image formats such as JPG, JPEG, and PNG. The library can convert image files to PDF before processing.
* Document: Document files such as DOC and DOCX. The library can convert document files to PDF before processing. (on-dev)
## Notes
* The library utilizes the pdfcpu-cli package to execute PDF-related commands. Ensure that it is installed and accessible in your environment.
* Make sure to handle any errors that may occur during the PDF processing operations.