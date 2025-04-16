package sglog

import "log/slog"

const vmoduleKey = "vmodule"

type vmoduleValue struct {
	name slog.Value
	lvar slog.LevelVar
}

func (v *vmoduleValue) LogValue() slog.Value {
	return v.name
}

// VModule creates a verbosity as a slog.Attr, which when present with a
// slog.Logger, controls the verbosity of the module specific log message as
// per it's current log level. Users can change a module's log level
// dynamically without effecting other module's log level.
func VModule(name string, level slog.Level) slog.Attr {
	value := &vmoduleValue{
		name: slog.StringValue(name),
	}
	value.lvar.Set(level)
	return slog.Any(vmoduleKey, value)
}

// SetVModuleLevel changes a vmodule attribute's level dynamically. Returns
// false if input attribute is not a vmodule attribute.
func SetVModuleLevel(a slog.Attr, l slog.Level) bool {
	if a.Key != vmoduleKey {
		return false
	}
	value, ok := a.Value.Any().(*vmoduleValue)
	if !ok {
		return false
	}
	value.lvar.Set(l)
	return true
}

// VModuleLevel retrieves a vmodule attribute's level. Returns (0, false) if
// input attribute is not a vmodule attribute.
func VModuleLevel(a slog.Attr) (slog.Level, bool) {
	if a.Key != vmoduleKey {
		return 0, false
	}
	value, ok := a.Value.Any().(*vmoduleValue)
	if !ok {
		return 0, false
	}
	return value.lvar.Level(), true
}
