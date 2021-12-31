package license

import (
	_ "embed"
)

// Header is the copyright and license header text meant to be used in
// in the CLI helper text.
var Header = `dinkur Copyright (C) 2021 Kalle Fagerberg

  License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>
  This program comes with ABSOLUTELY NO WARRANTY; for details run 'dinkur --license-w'.
  This is free software, and you are welcome to redistribute it
  under certain conditions; run 'dinkur --license-c' for details.
`

// Conditions is the full terms and conditions section of the GNU GPL 3.0
// license text.
//
//go:embed license_conditions.txt
var Conditions string

// Warranty is the warranty section of the GNU GPL 3.0 license text.
//
//go:embed license_warranty.txt
var Warranty string
