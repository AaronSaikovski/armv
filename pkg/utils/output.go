package utils

import (
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
)

// OutputSuccess prints a green success banner. The full report is written
// separately to the output file.
func OutputSuccess(respStatus string) {
	fmt.Println(aurora.Bold(aurora.Green("\n*****************************************************************")))
	fmt.Println(aurora.Bold(aurora.Green("*** SUCCESS - No Azure Resource Validation issues found. ***")))
	fmt.Println(aurora.Sprintf(aurora.Green("*** Response Status OK - %s ***"), aurora.Green(respStatus)))
	fmt.Println(aurora.Bold(aurora.Green("*****************************************************************")))
}

// OutputFailSummary prints a concise red banner when validation fails.
// Full error details live in the Markdown report file; this is the
// at-a-glance summary for the terminal.
func OutputFailSummary(errorCount int, topFailures []string) {
	fmt.Println(aurora.Bold(aurora.Red("\n*****************************************************************")))
	fmt.Println(aurora.Bold(aurora.Red("*** Validation FAILED ***")))
	fmt.Println(aurora.Red(fmt.Sprintf("*** %d resource(s) reported errors ***", errorCount)))
	if len(topFailures) > 0 {
		fmt.Println(aurora.Red(fmt.Sprintf("*** Top failures: %s ***", strings.Join(topFailures, ", "))))
	}
	fmt.Println(aurora.Bold(aurora.Red("*****************************************************************")))
}
