package taskflow_test

import (
	"encoding"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func Test_TFParams_String(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "1",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().String("x")

		assert.Equal(t, "1", got, "should return proper parameters")
	})
}

func Test_TFParams_Int_valid_dec(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "10",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assert.Equal(t, 10, got, "should return proper parameter value")
	})
}

func Test_TFParams_Int_valid_binary(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "0b10",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assert.Equal(t, 2, got, "should return proper parameter value")
	})
}

func Test_TFParams_Int_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Int("x")

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_Int_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Int("x")

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Bool_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "true",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Bool("x")

		assert.Equal(t, true, got, "should return proper parameter value")
	})
}

func Test_TFParams_Bool_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Bool("x")

		assert.Equal(t, false, got, "should return false as the default value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_Bool_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Bool("x")

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Float64_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "1.2",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Float64("x")

		assert.Equal(t, 1.2, got, "should return proper parameter value")
	})
}

func Test_TFParams_Float64_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Float64("x")

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_Float64_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Float64("x")

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Duration_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "1m",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Duration("x")

		assert.Equal(t, time.Minute, got, "should return proper parameter value")
	})
}

func Test_TFParams_Duration_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Duration("x")

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_Duration_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Duration("x")

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_Date_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "2000-03-05",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Date("x", "2006-01-02")

		assert.Equal(t, time.Date(2000, 3, 5, 0, 0, 0, 0, time.UTC), got, "should return proper parameter value")
	})
}

func Test_TFParams_Date_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		got := tf.Params().Date("x", "2006-01-02")

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_Date_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().Date("x", "2006-01-02")

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "2000-03-05T13:20:00Z",
		},
	}
	r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)

		assert.Equal(t, time.Date(2000, 3, 5, 13, 20, 0, 0, time.UTC), got, "should return proper parameter value")
	})
}

func Test_TFParams_ParseText_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_ParseText_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		var got time.Time
		tf.Params().ParseText("x", &got)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_nil(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got encoding.TextUnmarshaler
		tf.Params().ParseText("x", got)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseText_non_ptr(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got nonPtrTextUnmarshaler
		tf.Params().ParseText("x", got)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

type nonPtrTextUnmarshaler struct{}

func (nonPtrTextUnmarshaler) UnmarshalText([]byte) error {
	return nil
}

func Test_TFParams_ParseJSON_valid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": `x={ "A" : "abc" }`,
		},
	}
	r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)

		assert.Equal(t, x{A: "abc"}, got, "should return proper parameter value")
	})
}

func Test_TFParams_ParseJSON_missing(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)

		assert.Zero(t, got, "should return proper parameter value")
	})
	assert.True(t, result.Passed(), "the command should pass")
}

func Test_TFParams_ParseJSON_invalid(t *testing.T) {
	r := taskflow.Runner{
		Params: taskflow.Params{
			"x": "abc",
		},
	}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", &got)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseJSON_nil(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		tf.Params().ParseJSON("x", nil)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

func Test_TFParams_ParseJSON_non_ptr(t *testing.T) {
	r := taskflow.Runner{}
	result := r.Run(func(tf *taskflow.TF) {
		var got x
		tf.Params().ParseJSON("x", got)

		assert.Fail(t, "should not reach this line")
	})
	assert.True(t, result.Failed(), "the command should fail")
}

type x struct {
	A string
}
