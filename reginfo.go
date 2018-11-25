package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// RegInfo type
type RegInfo struct {
	nickname     string
	passwordHash []byte
}

// RegMap type
type RegMap map[string]*RegInfo

func (reg RegInfo) checkPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(reg.passwordHash, []byte(password)) == nil
}

func (m RegMap) createNewReg(n string, p string) (*RegInfo, bool) {

	hash, ok := hashAndSalt(p)
	if !ok {
		log.Println("Unable to set password hash for", n)
		return nil, false
	}

	m[n] = &RegInfo{nickname: n, passwordHash: hash}

	return m[n], true
}

func (m RegMap) isAlreadyRegistered(nickname string) bool {
	_, alreadyRegistered := m[nickname]
	return alreadyRegistered
}

func hashAndSalt(pwd string) ([]byte, bool) {

	bpwd := []byte(pwd)
	hash, err := bcrypt.GenerateFromPassword(bpwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
		return bpwd, false
	}
	return hash, true
}
