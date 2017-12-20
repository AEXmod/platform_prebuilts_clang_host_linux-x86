//
// Copyright (C) 2017 The Android Open Source Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package clangprebuilts

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/google/blueprint/proptools"

	"android/soong/android"
	"android/soong/cc"
	"android/soong/cc/config"
)

// This module is used to generate libfuzzer, libomp static libraries and
// libclang_rt.* shared libraries. When LLVM_PREBUILTS_VERSION and
// LLVM_RELEASE_VERSION are set, the library will generated from the given
// path.

func init() {
	android.RegisterModuleType("llvm_prebuilt_library_static",
		llvmPrebuiltLibraryStaticFactory)
	android.RegisterModuleType("libclang_rt_prebuilt_library_shared",
		libClangRtPrebuiltLibrarySharedFactory)
	android.RegisterModuleType("libclang_rt_prebuilt_library_static",
		libClangRtPrebuiltLibraryStaticFactory)
}

func getClangPrebuiltDir(ctx android.LoadHookContext) string {
	return path.Join(
		"./",
		ctx.AConfig().GetenvWithDefault("LLVM_PREBUILTS_VERSION", config.ClangDefaultVersion),
	)
}

func getClangResourceDir(ctx android.LoadHookContext) string {
	clangDir := getClangPrebuiltDir(ctx)
	releaseVersion := ctx.AConfig().GetenvWithDefault("LLVM_RELEASE_VERSION",
		config.ClangDefaultShortVersion)
	return path.Join(clangDir, "lib64", "clang", releaseVersion, "lib", "linux")
}

type archProps struct {
	Android_arm struct {
		Srcs []string
	}
	Android_arm64 struct {
		Srcs []string
	}
	Android_mips struct {
		Srcs []string
	}
	Android_mips64 struct {
		Srcs []string
	}
	Android_x86 struct {
		Srcs []string
	}
	Android_x86_64 struct {
		Srcs []string
	}
}

func llvmPrebuiltLibraryStatic(ctx android.LoadHookContext) {
	// Because of b/38393317, changing clang base dir is not allowed.  Mark
	// libFuzzer and libomp as disabled if LLVM_PREBUILTS_BASE is used to
	// specify a different base dir other than
	// $ANDROID_BUILD_TOP/prebuilts/clang/host (i.e. $CWD/..).  libFuzzer
	// would be unavailable only for the stage2 of the aosp-llvm build,
	// where it is not needed.
	var enableLLVMPrebuilts bool
	enableLLVMPrebuilts = true
	if prebuiltsBase := ctx.AConfig().Getenv("LLVM_PREBUILTS_BASE"); prebuiltsBase != "" {
		prebuiltsBaseAbs, err1 := filepath.Abs(prebuiltsBase)
		moduleBaseAbs, err2 := filepath.Abs("..")
		if err1 == nil && err2 == nil && prebuiltsBaseAbs != moduleBaseAbs {
			enableLLVMPrebuilts = false
		}
	}

	libDir := getClangResourceDir(ctx)
	name := strings.TrimPrefix(ctx.ModuleName(), "prebuilt_") + ".a"

	type props struct {
		Enabled             *bool
		Export_include_dirs []string
		Target              archProps
	}

	p := &props{}

	p.Enabled = proptools.BoolPtr(enableLLVMPrebuilts)
	if (name == "libFuzzer.a") {
		headerDir := path.Join(getClangPrebuiltDir(ctx), "prebuilt_include", "llvm", "lib", "Fuzzer")
		p.Export_include_dirs = []string{headerDir}
	}

	p.Target.Android_arm.Srcs = []string{path.Join(libDir, "arm", name)}
	p.Target.Android_arm64.Srcs = []string{path.Join(libDir, "aarch64", name)}
	p.Target.Android_mips.Srcs = []string{path.Join(libDir, "mips", name)}
	p.Target.Android_mips64.Srcs = []string{path.Join(libDir, "mips64", name)}
	p.Target.Android_x86.Srcs = []string{path.Join(libDir, "i386", name)}
	p.Target.Android_x86_64.Srcs = []string{path.Join(libDir, "x86_64", name)}
	ctx.AppendProperties(p)
}

func libClangRtPrebuiltLibraryShared(ctx android.LoadHookContext) {
	if ctx.AConfig().IsEnvTrue("FORCE_BUILD_SANITIZER_SHARED_OBJECTS") {
		return
	}

	libDir := getClangResourceDir(ctx)

	type props struct {
		Srcs []string
		System_shared_libs []string
		Sanitize struct {
			Never bool
		}
		Strip struct {
			None bool
		}
		Pack_relocations *bool
		Stl *string
	}

	p := &props{}

	name := strings.Replace(ctx.ModuleName(), "prebuilt_", "", 1)

	p.Srcs = []string{path.Join(libDir, name+".so")}
	p.System_shared_libs = []string{}
	p.Sanitize.Never = true
	p.Strip.None = true
	disable := false
	p.Pack_relocations = &disable
	none := "none"
	p.Stl = &none
	ctx.AppendProperties(p)
}

func libClangRtPrebuiltLibraryStatic(ctx android.LoadHookContext) {
	libDir := getClangResourceDir(ctx)

	type props struct {
		Srcs []string
	}

	name := strings.Replace(ctx.ModuleName(), "prebuilt_", "", 1)

	p := &props{}
	p.Srcs = []string{path.Join(libDir, name + ".a")}
	ctx.AppendProperties(p)
}

func llvmPrebuiltLibraryStaticFactory() android.Module {
	module, _ := cc.NewPrebuiltStaticLibrary(android.DeviceSupported)
	android.AddLoadHook(module, llvmPrebuiltLibraryStatic)
	return module.Init()
}

func libClangRtPrebuiltLibrarySharedFactory() android.Module {
	module, _ := cc.NewPrebuiltSharedLibrary(android.DeviceSupported)
	android.AddLoadHook(module, libClangRtPrebuiltLibraryShared)
	return module.Init()
}

func libClangRtPrebuiltLibraryStaticFactory() android.Module {
	module, _ := cc.NewPrebuiltStaticLibrary(android.DeviceSupported)
	android.AddLoadHook(module, libClangRtPrebuiltLibraryStatic)
	return module.Init()
}
