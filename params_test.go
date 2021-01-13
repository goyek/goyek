package taskflow_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func Test_Params_SetInt(t *testing.T) {
	params := taskflow.Params{}

	params.SetInt("x", 1)
	got := params["x"]

	assert.Equal(t, "1", got, "should return proper parameter value")
}

func Test_Params_SetBool(t *testing.T) {
	params := taskflow.Params{}

	params.SetBool("x", true)
	got := params["x"]

	assert.Equal(t, "true", got, "should return proper parameter value")
}

func Test_Params_SetDuration(t *testing.T) {
	params := taskflow.Params{}

	params.SetDuration("x", time.Second)
	got := params["x"]

	assert.Equal(t, "1s", got, "should return proper parameter value")
}

func Test_Params_SetDate(t *testing.T) {
	params := taskflow.Params{}

	params.SetDate("x", time.Date(2000, 3, 5, 0, 0, 0, 0, time.UTC), "2006-01-02")
	got := params["x"]

	assert.Equal(t, "2000-03-05", got, "should return proper parameter value")
}

func Test_Params_SetText_valid(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetText("x", time.Date(2000, 3, 5, 13, 20, 0, 0, time.UTC))
	got := params["x"]

	assert.NoError(t, err, "should not return any error")
	assert.Equal(t, "2000-03-05T13:20:00Z", got, "should return proper parameter value")
}

type badTextMarshaler struct{}

func (badTextMarshaler) MarshalText() ([]byte, error) {
	return nil, errors.New("failing")
}

func Test_Params_SetText_invalid(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetText("x", badTextMarshaler{})
	got := params["x"]

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_Params_SetText_nil(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetText("x", nil)
	got := params["x"]

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_Params_SetJSON_valid(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetJSON("x", x{A: "abc"})
	got := params["x"]

	assert.NoError(t, err, "should not return any error")
	assert.Equal(t, "{\"A\":\"abc\"}", got, "should return proper parameter value")
}

type badJSONMarshaler struct{}

func (badJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("failing")
}

func Test_Params_SetJSON_invalid(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetJSON("x", badJSONMarshaler{})
	got := params["x"]

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}

func Test_Params_SetJSON_nil(t *testing.T) {
	params := taskflow.Params{}

	err := params.SetJSON("x", nil)
	got := params["x"]

	assert.Error(t, err, "should tell that it failed to parse the value")
	assert.Zero(t, got, "should return proper parameter value")
}
