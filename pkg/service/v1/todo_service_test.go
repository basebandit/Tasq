package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	sqlMock "github.com/DATA-DOG/go-sqlmock"
	v1 "github.com/basebandit/go-grpc/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestToDoServiceServerCreate(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlMock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewToDoServiceServer(db)
	tm := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(tm)
	estimatedTimeOfCompletion, _ := ptypes.TimestampProto(tm)
	actualTimeOfCompletion, _ := ptypes.TimestampProto(tm)

	type args struct {
		ctx context.Context
		req *v1.CreateRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.CreateResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:                     "title",
						Description:               "description",
						Status:                    "status",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", "status", tm, tm, tm).WillReturnResult(sqlMock.NewResult(1, 1))
			},
			want: &v1.CreateResponse{
				Api: "v1",
				Id:  1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1000",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						EstimatedTimeOfCompletion: &timestamp.Timestamp{
							Seconds: 3,
							Nanos:   2,
						},
						ActualTimeOfCompletion: &timestamp.Timestamp{
							Seconds: 3,
							Nanos:   2,
						},
						Reminder: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "Invalid Reminder field format",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						EstimatedTimeOfCompletion: &timestamp.Timestamp{
							Seconds: 3,
							Nanos:   2,
						},
						Reminder: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "INSERT FAILED",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:                     "title",
						Description:               "desription",
						Status:                    "status",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", "status", tm, tm).WillReturnError(errors.New("INSERT failed"))
			},
			wantErr: true,
		},
		{
			name: "LastInsertId failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:                     "title",
						Description:               "description",
						Status:                    "status",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", "status", tm, tm).WillReturnResult(sqlMock.NewErrorResult(errors.New("LastInsertId failed")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Create(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Create() = %v, want %v", got, tt.want)
			}
		})
	}

}

