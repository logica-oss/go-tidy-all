package tidyall

// Exported for external tests.
type DirFilter = dirFilter

func (f DirFilter) Skip(path string, isDir bool) bool {
	return f.skip(path, isDir)
}

var (
	CompareGoVersion = compareGoVersion
	GoVersionPart    = goVersionPart
	FindModules      = findModules
	NewDirFilter     = newDirFilter

	PinGoDirective = pinGoDirective
	ReadGoVersion  = readGoVersion
	WriteGoVersion = writeGoVersion
	ParseModFile   = parseModFile
	Tidy           = tidy

	ValidateMinGoVersion  = validateMinGoVersion
	NormalizeMinGoVersion = normalizeMinGoVersion
)
