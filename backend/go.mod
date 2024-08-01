module go-authentication-boilerplate

go 1.15

replace go-authentication-boilerplate => ./

require (
	cloud.google.com/go/storage v1.43.0
	github.com/SherClockHolmes/webpush-go v1.3.0
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/bold-commerce/go-shopify/v4 v4.5.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gofiber/fiber/v2 v2.1.1
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.3.0
	github.com/resend/resend-go/v2 v2.9.0
	google.golang.org/api v0.189.0
	gorm.io/driver/postgres v1.0.5
	gorm.io/gorm v1.20.5
)
