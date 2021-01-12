package taskflow_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func Test_default_params(t *testing.T) {
	flow := taskflow.New()
	flow.Params["x"] = "1"
	flow.Params["z"] = "0"
	var got taskflow.Params
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf.Params()
		},
	})

	exitCode := flow.Run(context.Background(), "y=2", "z=3", "task")

	want := taskflow.Params{
		"x": "1",
		"y": "2",
		"z": "3",
	}
	assert.Equal(t, 0, exitCode, "should pass")
	assert.Equal(t, want, got, "should return proper parameters")
}

func Test_params(t *testing.T) {
	tf := testTF(t, "x=1")

	got := tf.Params()

	want := taskflow.Params{
		"x": "1",
	}
	assert.Equal(t, want, got, "should return proper parameters")
}

func Test_params_Int_valid_dec(t *testing.T) {
	tf := testTF(t, "x=10")

	got, err := tf.Params().Int("x")

	assert.NoError(t, err, "should parse the value")
	assert.Equal(t, 10, got, "should return proper parameter value")
}

func Test_params_Int_valid_binary(t *testing.T) {
	tf := testTF(t, "x=0b10")

	got, err := tf.Params().Int("x")

	assert.NoError(t, err, "should parse the value")
	assert.Equal(t, 2, got, "should return proper parameter value")
}

func Test_params_Int_missing(t *testing.T) {
	tf := testTF(t)

	got, err := tf.Params().Int("x")

	assert.NoError(t, err, "should not return any error")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Int_invalid(t *testing.T) {
	tf := testTF(t, "x=abc")

	got, err := tf.Params().Int("x")

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Bool_valid(t *testing.T) {
	tf := testTF(t, "x=true")

	got, err := tf.Params().Bool("x")

	assert.NoError(t, err, "should parse the value")
	assert.Equal(t, true, got, "should return proper parameter value")
}

func Test_params_Bool_missing(t *testing.T) {
	tf := testTF(t)

	got, err := tf.Params().Bool("x")

	assert.NoError(t, err, "should not return any error")
	assert.Equal(t, false, got, "should return false as the default value")
}

func Test_params_Bool_invalid(t *testing.T) {
	tf := testTF(t, "x=abc")

	got, err := tf.Params().Bool("x")

	assert.EqualError(t, err, "taskflow: parameter \"x\": strconv.ParseBool: parsing \"abc\": invalid syntax", "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Float64_valid(t *testing.T) {
	tf := testTF(t, "x=1.2")

	got, err := tf.Params().Float64("x")

	assert.NoError(t, err, "should parse the value")
	assert.Equal(t, 1.2, got, "should return proper parameter value")
}

func Test_params_Float64_missing(t *testing.T) {
	tf := testTF(t)

	got, err := tf.Params().Float64("x")

	assert.NoError(t, err, "should not return any error")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Float64_invalid(t *testing.T) {
	tf := testTF(t, "x=abc")

	got, err := tf.Params().Float64("x")

	var paramErr *strconv.NumError
	assert.True(t, errors.As(err, &paramErr), "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Duration_valid(t *testing.T) {
	tf := testTF(t, "x=1m")

	got, err := tf.Params().Duration("x")

	assert.NoError(t, err, "should parse the value")
	assert.Equal(t, time.Minute, got, "should return proper parameter value")
}

func Test_params_Duration_missing(t *testing.T) {
	tf := testTF(t)

	got, err := tf.Params().Duration("x")

	assert.NoError(t, err, "should not return any error")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_params_Duration_invalid(t *testing.T) {
	tf := testTF(t, "x=abc")

	got, err := tf.Params().Duration("x")

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}
