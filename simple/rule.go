package simple

type (
	Rule interface {
		Check(Value) bool
	}

	StringRule struct {
		Name      string
		MinLength int
		MaxLength int
	}

	IntRule struct {
		Name string
		Min  int
		Max  int
	}

	FloatRule struct {
		Name string
		Min  float32
		Max  float32
	}
)

/* -------------- rule -------- */

func (this *StringRule) Check(value Value) bool {
	if n := len(value.String(this.Name)); n < this.MinLength || n > this.MaxLength {
		return false
	}
	return true
}

func (this *IntRule) Check(value Value) bool {
	if n := value.Int(this.Name); n < this.Min || n > this.Max {
		return false
	}
	return true
}

func (this *FloatRule) Check(value Value) bool {
	if n := value.Float(this.Name); n < this.Min || n > this.Max {
		return false
	}
	return true
}

/* ---------- check --------------*/

func RuleCheck(value Value, rules ...Rule) bool {
	for _, r := range rules {
		if !r.Check(value) {
			return false
		}
	}

	return true
}
