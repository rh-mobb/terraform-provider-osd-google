/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	ocmlogging "github.com/openshift-online/ocm-sdk-go/logging"
)

// TfLogger is a logger that bridges OCM SDK logging to Terraform's tflog.
type TfLogger struct {
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
}

// New creates a logger that forwards OCM SDK logs to Terraform's logging.
func New() ocmlogging.Logger {
	return &TfLogger{
		debugEnabled: true,
		infoEnabled:  true,
		warnEnabled:  true,
		errorEnabled: true,
	}
}

// DebugEnabled returns true iff the debug level is enabled.
func (l *TfLogger) DebugEnabled() bool {
	return l.debugEnabled
}

// InfoEnabled returns true iff the information level is enabled.
func (l *TfLogger) InfoEnabled() bool {
	return l.infoEnabled
}

// WarnEnabled returns true iff the warning level is enabled.
func (l *TfLogger) WarnEnabled() bool {
	return l.warnEnabled
}

// ErrorEnabled returns true iff the error level is enabled.
func (l *TfLogger) ErrorEnabled() bool {
	return l.errorEnabled
}

// Debug sends to the log a debug message.
func (l *TfLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.debugEnabled {
		tflog.Debug(ctx, fmt.Sprintf(format, args...))
	}
}

// Info sends to the log an information message.
func (l *TfLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.infoEnabled {
		tflog.Info(ctx, fmt.Sprintf(format, args...))
	}
}

// Warn sends to the log a warning message.
func (l *TfLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.warnEnabled {
		tflog.Warn(ctx, fmt.Sprintf(format, args...))
	}
}

// Error sends to the log an error message.
func (l *TfLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.errorEnabled {
		tflog.Error(ctx, fmt.Sprintf(format, args...))
	}
}

// Fatal sends to the log an error message and exits.
func (l *TfLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	tflog.Error(ctx, fmt.Sprintf(format, args...))
	os.Exit(1)
}
