package main

const (
	exitOK int = iota
	exitBadCommandLine
	exitNoInputFiles
	exitErrorReadingFile
	exitErrorParsingSchema
	exitErrorGeneratingCode
	exitErrorWritingFile
)
