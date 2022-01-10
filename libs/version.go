package libs

import "fmt"

const (
    // VERSION of this project
    VERSION = "beta v4.0.1"
    // DESC description of the tool
    DESC = "A Workflow Engine for Offensive Security"
    // BINARY name of Osmedeus
    BINARY = "osmedeus"
    // SNAPSHOT binary name of Osmedeus
    SNAPSHOT = "osmp"
    // AUTHOR of this
    AUTHOR = "@j3ssiejjj"
    // DOCS private document
    DOCS = "https://docs.osmedeus.org/"
)

// TEMP default folder to store inputs
var TEMP = fmt.Sprintf("/tmp/%s-inputs/", SNAPSHOT)
