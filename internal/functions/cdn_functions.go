package functions

import (
	"context"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/storage"
	"go.uber.org/zap"
)

// cdnUpload uploads a file to cloud storage
// Usage: cdnUpload(localPath, remotePath) -> bool
func (vf *vmFunc) cdnUpload(call goja.FunctionCall) goja.Value {
	localPath := call.Argument(0).String()
	remotePath := call.Argument(1).String()
	logger.Get().Debug("Calling cdnUpload", zap.String("localPath", localPath), zap.String("remotePath", remotePath))

	if localPath == "undefined" || localPath == "" {
		logger.Get().Warn("cdnUpload: empty local path provided")
		return vf.vm.ToValue(false)
	}
	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnUpload: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.UploadFile(ctx, localPath, remotePath)
	if err != nil {
		logger.Get().Warn("cdnUpload: upload failed", zap.String("localPath", localPath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnUpload result", zap.String("localPath", localPath), zap.String("remotePath", remotePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// cdnDownload downloads a file from cloud storage
// Usage: cdnDownload(remotePath, localPath) -> bool
func (vf *vmFunc) cdnDownload(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	localPath := call.Argument(1).String()
	logger.Get().Debug("Calling cdnDownload", zap.String("remotePath", remotePath), zap.String("localPath", localPath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnDownload: empty remote path provided")
		return vf.vm.ToValue(false)
	}
	if localPath == "undefined" || localPath == "" {
		logger.Get().Warn("cdnDownload: empty local path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.DownloadFile(ctx, remotePath, localPath)
	if err != nil {
		logger.Get().Warn("cdnDownload: download failed", zap.String("remotePath", remotePath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnDownload result", zap.String("remotePath", remotePath), zap.String("localPath", localPath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// cdnExists checks if a file exists in cloud storage
// Usage: cdnExists(remotePath) -> bool
func (vf *vmFunc) cdnExists(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnExists", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnExists: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	client, err := storage.NewClientFromGlobal()
	if err != nil {
		logger.Get().Warn("cdnExists: failed to create storage client", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	exists, _ := client.Exists(ctx, remotePath)
	logger.Get().Debug("cdnExists result", zap.String("remotePath", remotePath), zap.Bool("exists", exists))
	return vf.vm.ToValue(exists)
}

// cdnDelete deletes a file from cloud storage
// Usage: cdnDelete(remotePath) -> bool
func (vf *vmFunc) cdnDelete(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnDelete", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnDelete: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.DeleteFile(ctx, remotePath)
	if err != nil {
		logger.Get().Warn("cdnDelete: delete failed", zap.String("remotePath", remotePath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnDelete result", zap.String("remotePath", remotePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}
