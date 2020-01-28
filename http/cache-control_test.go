package http

import "testing"

func TestParseCacheControl(t *testing.T) {
	dirs, err := ParseCacheControl("no-cache")
	if err != nil {
		t.Error("ParseCacheControl(\"no-cache\") error:", err)
		return
	}
	if len(dirs) != 1 {
		t.Error("no-cached parsed as", len(dirs), "directives")
		return
	}
	if dirs[0].Name != "no-cache" {
		t.Error("no-cached parsed with Name=", dirs[0].Name)
		return
	}
	if dirs[0].ArgGiven {
		t.Error("no-cached has an arg??")
		return
	}
	if dirs[0].Arg != "" {
		t.Error("no-cached has a non-empty arg??")
		return
	}

	dirs, err = ParseCacheControl("max-age=1")
	if err != nil {
		t.Error("ParseCacheControl(\"max-age=1\") error:", err)
		return
	}
	if len(dirs) != 1 {
		t.Error("max-age=1 parsed as", len(dirs), "directives")
		return
	}
	if dirs[0].Name != "max-age" {
		t.Error("max-age=1 parsed with Name=", dirs[0].Name)
		return
	}
	if !dirs[0].ArgGiven {
		t.Error("max-age=1 doesn't have an arg??")
		return
	}
	if dirs[0].Arg != "1" {
		t.Error("max-age=1 has incorrect arg:", dirs[0].Arg)
		return
	}

	dirs, err = ParseCacheControl("max-age=25")
	if err != nil {
		t.Error("ParseCacheControl(\"max-age=25\") error:", err)
		return
	}
	if len(dirs) != 1 {
		t.Error("max-age=25 parsed as", len(dirs), "directives")
		return
	}
	if dirs[0].Name != "max-age" {
		t.Error("max-age=25 parsed with Name=", dirs[0].Name)
		return
	}
	if !dirs[0].ArgGiven {
		t.Error("max-age=25 doesn't have an arg??")
		return
	}
	if dirs[0].Arg != "25" {
		t.Error("max-age=25 has incorrect arg:", dirs[0].Arg)
		return
	}

	dirs, err = ParseCacheControl("max-age=93451,private")
	if err != nil {
		t.Error("ParseCacheControl(\"max-age=93451,private\") error:", err)
		return
	}
	if len(dirs) != 2 {
		t.Error("max-age=93451,private parsed as", len(dirs), "directives")
		return
	}
	if dirs[0].Name != "max-age" {
		t.Error("max-age=93451 parsed with Name=", dirs[0].Name)
		return
	}
	if dirs[1].Name != "private" {
		t.Error("private parsed with Name=", dirs[1].Name)
		return
	}
	if !dirs[0].ArgGiven {
		t.Error("max-age=93451 doesn't have an arg??")
		return
	}
	if dirs[1].ArgGiven {
		t.Error("private has an arg??")
		return
	}
	if dirs[0].Arg != "93451" {
		t.Error("max-age=93451 has incorrect arg:", dirs[0].Arg)
		return
	}
	if dirs[1].Arg != "" {
		t.Error("private has non-empty arg:", dirs[1].Arg)
		return
	}

	dirs, err = ParseCacheControl("max-age=0   ,   private,public  ,no-transform,   oh=Hi,ah=\"0x4869\"")
	if err != nil {
		t.Error("ParseCacheControl(\"max-age=0   ,   private,public  ,no-transform,   oh=Hi,ah=\\\"0x4869\\\"\") error:", err)
		return
	}
	if len(dirs) != 6 {
		t.Error("max-age=0   ,   private,public  ,no-transform,   oh=Hi,ah=\"0x4869\" parsed as", len(dirs), "directives")
		return
	}
	if dirs[0].Name != "max-age" {
		t.Error("max-age=0 parsed with Name=", dirs[0].Name)
		return
	}
	if dirs[1].Name != "private" {
		t.Error("private parsed with Name=", dirs[1].Name)
		return
	}
	if dirs[2].Name != "public" {
		t.Error("public parsed with Name=", dirs[2].Name)
		return
	}
	if dirs[3].Name != "no-transform" {
		t.Error("no-transform parsed with Name=", dirs[3].Name)
		return
	}
	if dirs[4].Name != "oh" {
		t.Error("oh parsed with Name=", dirs[4].Name)
		return
	}
	if dirs[5].Name != "ah" {
		t.Error("ah parsed with Name=", dirs[5].Name)
		return
	}
	if !dirs[0].ArgGiven {
		t.Error("max-age=0 doesn't have an arg??")
		return
	}
	if dirs[1].ArgGiven {
		t.Error("private has an arg??")
		return
	}
	if dirs[2].ArgGiven {
		t.Error("public has an arg??")
		return
	}
	if dirs[3].ArgGiven {
		t.Error("no-transform has an arg??")
		return
	}
	if !dirs[4].ArgGiven {
		t.Error("oh hoesn't have an arg??")
		return
	}
	if !dirs[5].ArgGiven {
		t.Error("ah hoesn't have an arg??")
		return
	}
	if dirs[0].Arg != "93451" {
		t.Error("max-age=93451 has incorrect arg:", dirs[0].Arg)
		return
	}
	if dirs[1].Arg != "" {
		t.Error("private has non-empty arg:", dirs[1].Arg)
		return
	}
	if dirs[2].Arg != "" {
		t.Error("public has non-empty arg:", dirs[2].Arg)
		return
	}
	if dirs[3].Arg != "" {
		t.Error("no-transform has non-empty arg:", dirs[3].Arg)
		return
	}
	if dirs[4].Arg != "Hi" {
		t.Error("oh has incorrect arg:", dirs[4].Arg)
		return
	}
	if dirs[5].Arg != "0x4869" {
		t.Error("ah has incorrect arg:", dirs[5].Arg)
		return
	}
}
