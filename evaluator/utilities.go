package evaluator

// Unit2numeric converts a unit of length (cm, mm, pt, in, …)
// to internal scaled points.
//
// TODO complete this.
func Unit2numeric(u string) float64 {
	switch u {
	case "in":
		return 0.01388888
	}
	return 1.0
}

// ScaleDimension scales a numeric value by a unit.
func ScaleDimension(dimen float64, unit string) float64 {
	u := Unit2numeric(unit)
	return dimen * u
}
