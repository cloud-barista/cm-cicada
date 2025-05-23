package common

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func ReturnErrorMsg(c echo.Context, msg string) error {
	return c.JSONPretty(http.StatusBadRequest, ErrorResponse{Error: msg}, " ")
}

func ReturnInternalError(c echo.Context, err error, reason string) error {
	logger.Println(logger.ERROR, true, err.Error())

	msg := "Internal error occurred. (Reason: " + reason + ", Error: " + err.Error() + ")"

	return c.JSONPretty(http.StatusInternalServerError, ErrorResponse{Error: msg}, " ")
}

func ValidateTaskClearOptions(opt model.TaskClearOption) error {
	// only_failed 와 only_running 은 동시에 true면 안 됨
	if opt.OnlyFailed && opt.OnlyRunning {
		return errors.New("only_failed와 only_running은 동시에 true일 수 없습니다")
	}

	// dry_run 이 true인데 reset_dag_runs 도 true인 경우, 실제 reset은 발생하지 않음 (경고 또는 에러)
	if opt.DryRun && opt.ResetDagRuns {
		return errors.New("dry_run이 true이면 reset_dag_runs는 의미가 없습니다")
	}

	// include_future 가 true인데 only_failed 도 true면, 미래 태스크는 실패할 수 없음
	if opt.IncludeFuture && opt.OnlyFailed {
		return errors.New("include_future가 true이면 only_failed는 의미가 없습니다 (미래 태스크는 실패하지 않음)")
	}

	// 관련 태스크 포함 옵션이 모두 false이면, 처리할 수 있는 태스크가 거의 없음
	if !opt.IncludeUpstream && !opt.IncludeDownstream && !opt.IncludeSubdags && !opt.IncludeParentdag {
		fmt.Println("경고: include 관련 옵션이 모두 false이므로 단일 태스크만 처리됩니다. 이게 의도한 바인지 확인하세요.")
	}

	return nil
}
