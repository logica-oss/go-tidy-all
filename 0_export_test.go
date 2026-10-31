package tidyall

// Exported for external tests.
var (
	CompareGoVersion = compareGoVersion
	GoVersionPart    = goVersionPart
	FindModules      = findModules
	NewDirFilter     = newDirFilter
	CopyDirValidated = copyDirValidated

	_              = DirFilter{}
	PinGoDirective = pinGoDirective
	ReadGoVersion  = readGoVersion
	WriteGoVersion = writeGoVersion
	ParseModFile   = parseModFile
	CopyDir        = copyDir
	StageModule    = stageModule
	ProcessModule  = processModule
	Tidy           = tidy
)
