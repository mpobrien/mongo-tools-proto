package mongotop

import (
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/mongotop/options"
	"github.com/shelman/mongo-tools-proto/mongotop/output"
	"github.com/shelman/mongo-tools-proto/mongotop/result"
	"time"
)

// Wrapper for the mongotop functionality
type MongoTop struct {
	// generic mongo tool options
	Options *commonopts.MongoToolOptions

	// mongotop-specific options
	TopOptions *options.MongoTopOptions

	// for connecting to the db
	SessionProvider *db.SessionProvider

	// for outputting the results
	output.Outputter
}

// Connect to the database and spin, running the top command and outputting
// the results appropriately.
func (self *MongoTop) Run() error {

	// the top results from the previous run, used for diffing
	previousResults, err := self.runTopCommand()
	if err != nil {
		return fmt.Errorf("error running initial top command: %v", err)
	}

	for {

		// run the top command against the database
		topResults, err := self.runTopCommand()
		if err != nil {
			return fmt.Errorf("error running top command: %v", err)
		}

		// diff the results
		diff := result.Diff(previousResults, topResults)

		// output the results
		if err := self.Outputter.Output(diff, self.Options); err != nil {
			return fmt.Errorf("error outputting results: %v", err)
		}

		// update the previous results
		previousResults = topResults

		// sleep
		time.Sleep(time.Duration(self.TopOptions.SleepTime) * time.Second)
	}

	return nil
}

// Run the top command against the database, and return the results.
func (self *MongoTop) runTopCommand() (*result.TopResults, error) {

	// get a database session
	session, err := self.SessionProvider.GetSession()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database server: %v", err)
	}
	defer session.Close()

	// get the admin database
	adminDB := session.DB("admin")
	res := &result.TopResults{}

	// run the command
	if err := adminDB.Run("top", res); err != nil {
		return nil, fmt.Errorf("error running top command: %v", err)
	}

	// success
	return res, nil

}
