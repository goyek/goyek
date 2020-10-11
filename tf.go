package taskflow

import "context"

// TF TODO.
type TF struct {
	ctx context.Context
}

// Context TODO.
func (tf *TF) Context() context.Context {
	return tf.ctx
}
