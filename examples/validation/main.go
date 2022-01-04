// Example program for validation, showcasing the following:
// Validation of parameters as a dependent, reusable task.
// Execute `go run . -v migrate -db only-up -direction up` as a valid example.
// Execute `go run . -v migrate -db only-up -direction down` as an invalid example.
// Execute `go run . -h` to see all details.

package main

import (
	"github.com/goyek/goyek"
)

func main() {
	flow := &goyek.Flow{}

	db := flow.RegisterStringParam(goyek.StringParam{
		Name:  "db",
		Usage: "The database to migrate ('only-up' / 'only-down')",
	})
	direction := flow.RegisterStringParam(goyek.StringParam{
		Name:  "direction",
		Usage: "The direction to migrate ('up' / 'down')",
	})

	flow.Register(taskMigrate(db, direction))

	flow.Main()
}

// validateMigrateParams validates params and prints error messages for each error
// it encounters. If it encounters at least one error, the validation function
// will abort the TaskFlow and the task will be stopped.
func validateMigrateParams(tf *goyek.TF, db, direction goyek.RegisteredStringParam) {
	var err bool
	dbStr := db.Get(tf)
	dirStr := direction.Get(tf)
	if dbStr != "only-up" && dbStr != "only-down" {
		tf.Errorf("Database %q is invalid", dbStr)
		err = true
	}
	if dirStr != "up" && dirStr != "down" {
		tf.Errorf("Direction %q is invalid", dirStr)
		err = true
	}
	if (dbStr == "only-up" && dirStr == "down") || (dbStr == "only-down" && dirStr == "up") {
		tf.Errorf("Direction %q is invalid for database %q", dirStr, dbStr)
		err = true
	}
	if err {
		tf.FailNow()
	}
}

// taskMigrate validates the params for the task, and if they are valid,
// it prints a message with the provided params.
func taskMigrate(db, direction goyek.RegisteredStringParam) goyek.Task {
	return goyek.Task{
		Name:   "migrate",
		Usage:  "Migrate a database",
		Params: goyek.Params{db, direction},
		Action: func(tf *goyek.TF) {
			validateMigrateParams(tf, db, direction)
			tf.Logf("Database %q and direction %q are valid values", db.Get(tf), direction.Get(tf))
		},
	}
}
