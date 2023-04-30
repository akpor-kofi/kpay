package main

//
//import (
//	"fraud-detect-system/domain"
//	"github.com/jaswdr/faker"
//	"github.com/kamva/mgm/v3"
//	"time"
//)
//
//func mainx() {
//
//	// first fill the user model
//	fake := faker.New()
//
//	user := new(domain.User)
//	err := mgm.Coll(user).FindByID("63b399e1739d74c2dc404ea2", user)
//	check(err)
//
//	for j := 0; j < 2; j++ {
//		transaction := &domain.Transaction{
//			ID:       fake.UUID().V4(),
//			Amt:      1000,
//			Category: fake.Beer().Alcohol(),
//			UnixTime: time.Now(),
//			IsFraud:  false,
//			User:     *user,
//			Merchant: domain.Merchant{
//				Name:      fake.Person().Name(),
//				MerchLat:  fake.Address().Latitude(),
//				MerchLong: fake.Address().Longitude(),
//			},
//		}
//
//		err := mgm.Coll(transaction).Create(transaction)
//		check(err)
//
//		user.Balance -= transaction.Amt
//
//		err = mgm.Coll(user).Update(user)
//		check(err)
//
//	}
//}
