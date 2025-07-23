package tool

import "testing"

func TestValidateMobile(t *testing.T) {
	type args struct {
		mobile string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// 标准格式
		{"valid standard", args{"13800138000"}, true},
		{"valid 4G", args{"17600001111"}, true},
		{"valid 5G", args{"19900001111"}, true},

		// 带国家码
		{"valid with +86", args{"+8613900139000"}, true},
		{"valid with 86", args{"8613700137000"}, true},

		// 带分隔符
		{"valid with space", args{"136 0013 6000"}, true},
		{"valid with hyphen", args{"135-0013-5000"}, true},
		{"valid complex", args{"+86 134 0013 4000"}, true},

		// 无效格式
		{"invalid not start with 1", args{"23456789012"}, false},
		{"invalid too short", args{"12345"}, false},
		{"invalid too long", args{"131000011111"}, false},
		{"invalid letter", args{"132000a1111"}, false},
		{"invalid special char", args{"133000@1111"}, false},
		{"invalid country code", args{"+85213900001111"}, false}, // 香港区号
		{"invalid empty", args{""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateMobile(tt.args.mobile); got != tt.want {
				t.Errorf("ValidateMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}
