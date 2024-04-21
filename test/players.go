package test

import (
	"intraclub/model"
)

var Tom = &model.Player{UserId: TomEasum.ID.Hex(), Line: 1}
var Ethan = &model.Player{UserId: EthanMoland.ID.Hex(), Line: 1}
var Andy = &model.Player{UserId: AndyLascik.ID.Hex(), Line: 2}
var JD = &model.Player{UserId: JdArthur.ID.Hex(), Line: 2}
var Chris = &model.Player{UserId: ChrisBoehm.ID.Hex(), Line: 3}
var Norm = &model.Player{UserId: NormTaffet.ID.Hex(), Line: 3}
var Tomer = &model.Player{UserId: TomerWagshal.ID.Hex(), Line: 2}
var Paul = &model.Player{UserId: PaulCohen.ID.Hex(), Line: 2}
var Kevin = &model.Player{UserId: KevinCampbell.ID.Hex(), Line: 2}

func PlayerFromUserId(userId string) *model.Player {
	switch userId {
	case TomEasum.ID.Hex():
		return Tom
	case EthanMoland.ID.Hex():
		return Ethan
	case AndyLascik.ID.Hex():
		return Andy
	case JdArthur.ID.Hex():
		return JD
	case ChrisBoehm.ID.Hex():
		return Chris
	case NormTaffet.ID.Hex():
		return Norm
	case TomerWagshal.ID.Hex():
		return Tomer
	case PaulCohen.ID.Hex():
		return Paul
	case KevinCampbell.ID.Hex():
		return Kevin
	default:
		panic("unexpected user ID")
	}
}
