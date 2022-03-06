package codes

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Acceptable checks if given error is acceptable.
// 熔断：用来判断哪些error会计入失败计数
func Acceptable(err error) bool {
	switch status.Code(err) {
	// 异常请求错误
	case codes.DeadlineExceeded, codes.Internal, codes.Unavailable, codes.DataLoss:
		return false
	default:
		return true
	}
}
