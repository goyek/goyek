package taskflow_test

import (
	"encoding"
	"testing"
	"time"

	"github.com/pellared/taskflow"
)

func Test_TFParams_String(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "1",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().String("x")

		assertEqual(t, got, "1", "should return proper parameters")
	})
}

func Test_TFParams_Int_valid_dec(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "10",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assertEqual(t, got, 10, "should return proper parameter value")
	})
}

func Test_TFParams_Int_valid_binary(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "0b10",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assertEqual(t, got, 2, "should return proper parameter value")
	})
}

func Test_TFParams_Int_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assertEqual(t, got, 0, "should return proper parameter value")
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_Int_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Int("x")

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Bool_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "true",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Bool("x")

		assertEqual(t, got, true, "should return proper parameter value")
	})
}

func Test_TFParams_Bool_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Bool("x")
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_Bool_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Bool("x")

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Float64_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "1.2",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Float64("x")

		assertEqual(t, got, 1.2, "should return proper parameter value")
	})
}

func Test_TFParams_Float64_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Float64("x")
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_Float64_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Float64("x")

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Duration_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "1m",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Duration("x")

		assertEqual(t, got, time.Minute, "should return proper parameter value")
	})
}

func Test_TFParams_Duration_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Duration("x")
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_Duration_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Duration("x")

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Date_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "2000-03-05",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Date("x", "2006-01-02")

		assertEqual(t, got, time.Date(2000, 3, 5, 0, 0, 0, 0, time.UTC), "should return proper parameter value")
	})
}

func Test_TFParams_Date_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Date("x", "2006-01-02")
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_Date_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Date("x", "2006-01-02")

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "2000-03-05T13:20:00Z",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)

		assertEqual(t, got, time.Date(2000, 3, 5, 13, 20, 0, 0, time.UTC), "should return proper parameter value")
	})
}

func Test_TFParams_ParseText_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_ParseText_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_nil(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got encoding.TextUnmarshaler
		tf.Params().ParseText("x", got)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_non_ptr(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got nonPtrTextUnmarshaler
		tf.Params().ParseText("x", got)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

type nonPtrTextUnmarshaler struct{}

func (nonPtrTextUnmarshaler) UnmarshalText([]byte) error {
	return nil
}

func Test_TFParams_ParseJSON_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": `x={ "A" : "abc" }`,
		},
	}
	r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)

		assertEqual(t, got, x{A: "abc"}, "should return proper parameter value")
	})
}

func Test_TFParams_ParseJSON_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)
	})
	assertFalse(t, result.Passed(), "the command should fail")
}

func Test_TFParams_ParseJSON_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: map[string]string{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseJSON_nil(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().ParseJSON("x", nil)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseJSON_non_ptr(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", got)

		t.Error("should not reach this line")
	})
	assertTrue(t, result.Failed(), "the command should fail")
}

type x struct {
	A string
}
