package runtime

import (
	"reflect"
	"testing"

	env "github.com/ta2gch/iris/runtime/environment"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

func TestTagbody(t *testing.T) {
	local, global := env.New(), env.TopLevel
	defmacro("TAGBODY", tagbody)
	defmacro("GO", tagbodyGo)
	defmacro("CATCH", catch)
	defmacro("THROW", throw)
	defmacro("QUOTE", quote)
	type args struct {
		local  *env.Environment
		global *env.Environment
		obj    ilos.Instance
	}
	tests := []struct {
		name    string
		args    args
		want    ilos.Instance
		wantErr bool
	}{
		{
			name:    "tagbody & go",
			args:    args{local, global, readFromString("(catch 'foo (tagbody (go bar) (throw 'foo 1) bar))")},
			want:    instance.New(class.Null),
			wantErr: false,
		},
		{
			name:    "nested tagbody & go",
			args:    args{local, global, readFromString("(catch 'foo (tagbody (tagbody (go bar) (throw 'foo 1) bar (go foobar)) foobar))")},
			want:    instance.New(class.Null),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(tt.args.local, tt.args.global, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}