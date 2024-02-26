package test

import "intraclub/models"

// dummy IDs for team 1 and team 2

var Team1Id = "a2217507bc9c44ca8dda2bc224194ccd"
var Team2Id = "b4feb57989a846648cda083d97f7a913"

// some dummy Player records for 1s, 2s and 3s

var TomEasum = models.Player{
	PlayerId:  "4558a12ede0e4f759d88afe3d7dc469e",
	FirstName: "Tom",
	LastName:  "Easum",
	Line:      1,
}

var EthanMoland = models.Player{
	PlayerId:  "91ffa9782db9456bac7a37ec6f8b4569",
	FirstName: "Ethan",
	LastName:  "Moland",
	Line:      1,
}

var AndyLascik = models.Player{
	PlayerId:  "eb958972d1394e3383bb4ed1a9f9e5e1",
	FirstName: "Andy",
	LastName:  "Lascik",
	Line:      2,
}

var JdArthur = models.Player{
	PlayerId:  "d72b265080384b7a94ccdefa6b493e92",
	FirstName: "JD",
	LastName:  "Arthur",
	Line:      2,
}

var ChrisBoehm = models.Player{
	PlayerId:  "83a8ff71835948419b1495a1ddbb4d2f",
	FirstName: "Chris",
	LastName:  "Boehm",
	Line:      3,
}

var NormTaffet = models.Player{
	PlayerId:  "5fb9f643b308478b9854-8b811c1d353e",
	FirstName: "Norm",
	LastName:  "Taffet",
	Line:      3,
}
