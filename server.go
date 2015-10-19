package main

import (
	"github.com/go-martini/martini"
	"github.com/llun/martini-amber"
	"github.com/martini-contrib/sessions"
	"github.com/garyburd/redigo/redis"
	"time"
	"net/http"
	"log"
)

type Article struct{
	Title,Image,Body,Author string
}
var arts = []Article{
	{"Samsung Gear S2 chính thức: mặt tròn, viền xoay, có bản 3G độc lập, pin 2-3 ngày","https://1anh.com/500/0/EZYwYm5NdZIwlXo48tMBufykgNFiWK-L-tezFAefrD8HNAyJZL7vBA0WR1ECJ3M-bUupVuQOI5IT-pZ1Eo-Q_OOURoj6_OT-H9y9iHTvFzNfaDpTHhAYZSj27EoluslL","Hôm nay Samsung đã chính thức ra mắt những chiếc đồng hồ mới nhất của mình với tên gọi Gear S2. So với Gear S đời đầu thì thiết bị mới có kiểu dáng truyền thống hơn, không còn làm kiểu vòng đeo tay nữa. Có 3 phiên bản Gear S được tung ra: Gear S2 Classic với lớp hoàn thiện màu đen kèm dây da, Gear S2 Classic 3G với khả năng gọi điện, nhắn tin, kết nối mạng di động mà không cần điện thoại","secmák"},
	{"Đã có thể kết nối Android Wear với iOS từ 8.2 trở lên, hỗ trợ Google Now, có ứng dụng tải về","https://1anh.com/500/0/EZYwYm5NdZIwlXo48tMBufykgNFiWK-L-tezFAefrD_s-1jBU-WTCAm4DqGoMCjMp2SUsDWnj76rCqIkLEUt52p7Mtw-iK_jYy5pSsX_lvmeJRZUx9n_vzMpxa8hgH5b","Không cần đến ứng dụng bên thứ ba nữa, Google đã chính thức công bố việc hỗ trợ Android Wear dành cho nền tảng iOS phiên bản từ 8.2 trở lên, trước mắt là nó sẽ hoạt động được trên mẫu đồng hồ LG Watch Urbane. Như vậy người dùng iPhone 5 trở lên đã có thể kết nối đến các thiết bị Android Wear, trong tương lai sẽ được mở rộng với nhiều sản phẩm của các tên tuổi khác như Huawei, Asus và Motorola.","secmák"},
	{"250.000 tài khoản iCloud bị hack vì Jailbreak, khả năng bị khóa máy từ xa, mất tiền trên iTunes...","https://1anh.com/500/0/EZYwYm5NdZIwlXo48tMBufykgNFiWK-L-tezFAefrD_fn4Lzk9oxv4pOvbYUwOOJyLbwqqFdMhn024hMgx3N_w","Hãng nghiên cứu Palo Alto Networks cho biết đã có hơn 250.000 tài khoản iTunes bị mất cắp sau khi Jailbreak máy và cài trúng những malware chứa trong kho app Cydia. Những máy này không những tự động gửi password iTunes về cho hacker mà còn có thể bị mua đồ trong App Store mà người dùng không hề hay biết. Ngoài ra, những máy bị nhiễm còn có khả năng bị khóa máy từ xa để đòi tiền chuộc. Chỉ những máy đã Jailbreak mới có khả năng dính malware này","secmák"},
	{`Chiếc Nexus 5,2" sẽ có giá bán "gần với các máy cao cấp, có thể không mang tên Nexus 5?`,"https://1anh.com/500/0/EZYwYm5NdZIwlXo48tMBufykgNFiWK-L-tezFAefrD91hUHwUbkjt-XhPrK7lJ8QSvH2QBYo8VeBK-J9cCj9bQ",`Trang Android Police vừa nhận được thông tin từ một nguồn đáng tin cậy nói rằng chiếc Nexus màn hình 5,2" do LG sản xuất sẽ được bán ra vào cuối năm nay với giá "gần với các máy cao cấp". Người này còn nói rằng có thể "Nexus 5" không phải là tên chính thức, nhưng để cho gọn thì cứ tạm gọi là "Nexus 5 (2015)." Trước đây chúng ta đã biết rằng thiết bị này sử dụng tấm nền Full-HD, vi xử lý Qualcomm Snapdragon`,"secmák"},
	{"So sánh hình ảnh Hyundai Tucson 2016 và Honda CR-V 2015","https://1anh.com/500/0/EZYwYm5NdZIwlXo48tMBufykgNFiWK-L-tezFAefrD_g9Qzo0SSHgOuUG53Jw8qehjGp5TF1r3gXKGf2ZFFyEg",`Hyundai Tucson 2016 nhập khẩu trực tiếp từ Hàn Quốc là chiếc xe mới gia nhập phân khúc crossover cỡ nhỏ 5 chỗ tại Việt Nam. Phân khúc này trước đây thống trị bởi 2 chiếc xe lắp ráp trong nước là Honda CR-V và Mazda CX-5. Vì CX-5 sắp nhận đợt phân cấp facelift nên để cho công bằng chúng ta sẽ so sánh Hyundai Tucson 2016 vừa được giới thiệu trong tháng 8 với Honda CR-V 2015 ra mắt vào tháng 11 năm ngoái. Ở Việt Nam,`,"secmák"},
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(password) > 0{
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
type Vars struct{
	Category map[string]string
	Arts []Article
}
var rPool = newPool("localhost:6379","")

func main() {

	m := martini.Classic()
	m.Use(martini_amber.Renderer(map[string]string{

	}))
	store := sessions.NewCookieStore([]byte("secret...."))
	m.Use(sessions.Sessions("msession",store))
	m.Get("/", func(r martini_amber.Render,request *http.Request,session sessions.Session) {
		log.Println(request.UserAgent())
		name := session.Get("name")
		if name==nil && request.FormValue("name")!=""{
			session.Set("name",request.FormValue("name"))
		}
		vars := map[string]interface{}{"cats": []string{"Home", "Product", "Contact","Live Showing"},"arts":arts,"name":name}
		r.Amber(200, "home", vars)
	})
	m.Get("/test", func(r martini_amber.Render,request *http.Request,session sessions.Session) {
		log.Println(request.UserAgent())
		name := session.Get("name")
		if name==nil && request.FormValue("name")!=""{
			session.Set("name",request.FormValue("name"))
		}
		vars := map[string]interface{}{"cats": []string{"Home", "Product", "Contact","Live Showing"},"arts":arts,"name":name}
		r.Amber(200, "test", vars)
	})

	m.RunOnAddr(":8080")
}
