package tests

import (
	"testing"
	"xTools/logx"
)

func TestLog(t *testing.T) {
	logger := logx.NewLogger("debug", 1, "./logs", 7, true, false)

	logger.Info("user", "注册模块", "新用户注册成功")
	logger.Debug("order", "订单模块", "订单号生成中")
	logger.Error("order", "订单模块", "库存不足")

}
