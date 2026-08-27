package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type testComparePassword struct {
	name     string
	password string
	wantErr  bool
}

func TestComparePassword(t *testing.T) {
	password := "whoawicked"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	tests := []testComparePassword{
		{"correct password matches", password, false},
		{"wrong password fails", "wrongpassword", true},
		{"empty password fails", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(string(hash), tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestGenerateAndVerifyToken(t *testing.T) {
	secret := []byte("secret-key")
	email := "test@gmail.com"
	userId := "user-001"
	token, err := GenerateToken(email, userId, secret, 2)
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}
	claims, err := VerifyToken(token, secret)
	if err != nil {
		t.Fatalf("VerifyToken() failed: %v", err)
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}
	if claims.UserId != userId {
		t.Errorf("UserID = %v, want %v", claims.UserId, userId)
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	secret := []byte("correct-secret")
	wrongSecret := []byte("wrong-secret")
	token, _ := GenerateToken("test@gmail.com", "user-001", secret, 2)
	_, err := VerifyToken(token, wrongSecret)
	if err == nil {
		t.Error("VerifyToken() should fail with wrong secret, but got nil error")
	}
}
