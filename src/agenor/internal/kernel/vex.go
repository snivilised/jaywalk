package kernel

type (
	vexation interface {
		ancestor() string
		vapour() inspection
		cause() string
		magnitude() string
	}

	vex struct {
		data     any
		anc      string
		vap      inspection
		catalyst string
		mag      string
	}
)

func (v *vex) Data() any {
	return v.data
}

func (v *vex) ancestor() string {
	return v.anc
}

func (v *vex) vapour() inspection {
	return v.vap
}

func (v *vex) cause() string {
	return v.catalyst
}

func (v *vex) magnitude() string {
	return v.mag
}
