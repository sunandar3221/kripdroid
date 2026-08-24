//go:build android

package ui

/*
#cgo LDFLAGS: -landroid

#include <jni.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

var (
	bioMu       sync.Mutex
	bioRequests = make(map[int]chan error)
	nextBioID   int
)

func NativeBiometricPrompt(title, desc string) error {
	bioMu.Lock()
	id := nextBioID
	nextBioID++
	ch := make(chan error, 1)
	bioRequests[id] = ch
	bioMu.Unlock()

	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "com/kripdroid/app/BiometricHelper")
		if err != nil {
			return err
		}

		showMethod := jni.GetStaticMethodID(env, class, "show", "(Landroid/content/Context;Ljava/lang/String;Ljava/lang/String;I)V")
		
		err = jni.CallStaticVoidMethod(env, class, showMethod,
			jni.Value(app.AppContext()),
			jni.Value(jni.JavaString(env, title)),
			jni.Value(jni.JavaString(env, desc)),
			jni.Value(id),
		)
		return err
	})

	if err != nil {
		bioMu.Lock()
		delete(bioRequests, id)
		bioMu.Unlock()
		return err
	}

	return <-ch
}

//export Java_com_kripdroid_app_BiometricHelper_onSucceeded
func Java_com_kripdroid_app_BiometricHelper_onSucceeded(env *C.JNIEnv, _ C.jclass, id C.jint) {
	bioMu.Lock()
	if ch, ok := bioRequests[int(id)]; ok {
		ch <- nil
		delete(bioRequests, int(id))
	}
	bioMu.Unlock()
}

//export Java_com_kripdroid_app_BiometricHelper_onError
func Java_com_kripdroid_app_BiometricHelper_onError(env *C.JNIEnv, _ C.jclass, id C.jint, errStr C.jstring) {
	errText := ""
	if errStr != 0 {
		jen := jni.EnvFor(uintptr(unsafe.Pointer(env)))
		errText = jni.GoString(jen, jni.String(uintptr(errStr)))
	}
	
	bioMu.Lock()
	if ch, ok := bioRequests[int(id)]; ok {
		ch <- errors.New(errText)
		delete(bioRequests, int(id))
	}
	bioMu.Unlock()
}
