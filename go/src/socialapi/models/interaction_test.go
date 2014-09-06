package models

import (
	"socialapi/workers/common/runner"
	"testing"

	"github.com/koding/bongo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInteractiongetAccountId(t *testing.T) {
	r := runner.New("test")
	if err := r.Init(); err != nil {
		t.Fatalf("couldnt start bongo %s", err.Error())
	}
	defer r.Close()

	Convey("while getting account id", t, func() {
		Convey("it should have error if interaction id is not set", func() {
			i := NewInteraction()

			in, err := i.getAccountId()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "couldnt find accountId from content")
			So(in, ShouldEqual, 0)
		})

		Convey("it should have error if account is not set in db", func() {
			i := NewInteraction()
			i.Id = 4590

			in, err := i.getAccountId()
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, bongo.RecordNotFound)
			So(in, ShouldEqual, 0)
		})

		Convey("it should return quickly account id if id is set", func() {
			i := NewInteraction()
			i.AccountId = 1020

			in, err := i.getAccountId()
			So(err, ShouldBeNil)
			So(in, ShouldEqual, i.AccountId)
		})
	})
}

func TestInteractionisExempt(t *testing.T) {
	r := runner.New("test")
	if err := r.Init(); err != nil {
		t.Fatalf("couldnt start bongo %s", err.Error())
	}
	defer r.Close()

	Convey("While testing interaction is exempt or not", t, func() {
		Convey("it should have error while getting account id from db when channel id is not set", func() {
			i := NewInteraction()

			ie, err := i.isExempt()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "couldnt find accountId from content")
			So(ie, ShouldEqual, false)
		})

		Convey("it should have error if account id is not set", func() {
			i := NewInteraction()
			i.Id = 1098

			ie, err := i.isExempt()
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, bongo.RecordNotFound)
			So(ie, ShouldEqual, false)
		})

	})
}
