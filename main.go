package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	l.Warnf("hello")
}
