package errcode

// Code 业务错误码。
type Code int

const (
	CodeOK              Code = 0
	CodeInvalidParam    Code = 1001
	CodeRoomNotFound    Code = 1002
	CodeCatalogNotFound Code = 1003
	CodeNoRoute         Code = 2001
	CodeInternal        Code = 5000
)

// HTTPStatus 返回错误码对应的 HTTP 状态码。
func HTTPStatus(code Code) int {
	switch code {
	case CodeOK:
		return 200
	case CodeInvalidParam:
		return 400
	case CodeRoomNotFound:
		return 404
	case CodeCatalogNotFound:
		return 404
	case CodeNoRoute:
		return 422
	default:
		return 500
	}
}

// Message 返回错误码的默认文案。
func Message(code Code) string {
	switch code {
	case CodeOK:
		return "ok"
	case CodeInvalidParam:
		return "参数校验失败"
	case CodeRoomNotFound:
		return "房间不存在"
	case CodeCatalogNotFound:
		return "楼栋或楼层不存在"
	case CodeNoRoute:
		return "未找到可行路线"
	default:
		return "内部错误"
	}
}
