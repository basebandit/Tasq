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

func TestToDoServiceServerRead(t *testing.T) {
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
		req *v1.ReadRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.ReadResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				rows := sqlMock.NewRows([]string{"ID", "Title", "Description", "Status", "EstimatedTimeOfCompletion", "ActualTimeOfCompletion", "Reminder"}).AddRow(1, "title", "description", "status", tm, tm, tm)
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).WillReturnRows(rows)
			},
			want: &v1.ReadResponse{
				Api: "v1",
				ToDo: &v1.ToDo{
					Id:                        1,
					Title:                     "title",
					Description:               "description",
					Status:                    "status",
					EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
					ActualTimeOfCompletion:    actualTimeOfCompletion,
					Reminder:                  reminder,
				},
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "V1",
					Id:  1,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "SELECT failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).WillReturnError(errors.New("SELECT failed"))
			},
			wantErr: true,
		},
		{
			name: "Not found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				rows := sqlMock.NewRows([]string{"ID", "Title", "Description", "Status", "EstimatedTimeOfCompletion", "ActualTimeOfCompletion", "Reminder"})
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Read(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoserviceServer.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Read() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDoServiceUpdate(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlMock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	tm := time.Now().In(time.UTC)
	atc := time.Date(2020, 2, 27, 3, 15, 45, 34567, time.UTC)
	reminder, _ := ptypes.TimestampProto(tm)
	estimatedTimeOfCompletion, _ := ptypes.TimestampProto(tm)
	actualTimeOfCompletion, _ := ptypes.TimestampProto(atc)

	type args struct {
		ctx context.Context
		req *v1.UpdateRequest
	}

	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.UpdateResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:                        1,
						Title:                     "new title",
						Description:               "new description",
						Status:                    "Completed",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", "Completed", tm, atc, tm, 1).WillReturnResult(sqlMock.NewResult(1, 1))
			},
			want: &v1.UpdateResponse{
				Api:     apiVersion,
				Updated: 1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:                        1,
						Title:                     "new title",
						Description:               "new description",
						Status:                    "Started",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
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
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Status:      "Completed",
						EstimatedTimeOfCompletion: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
						ActualTimeOfCompletion: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
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
			name: "UPDATE failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:                        1,
						Title:                     "new title",
						Description:               "new description",
						Status:                    "Completed",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", "Completed", tm, atc, tm, 1).WillReturnError(errors.New("UPDATE failed"))
			},
			wantErr: true,
		},
		{
			name: "RowsAffected failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:                        1,
						Title:                     "new title",
						Description:               "new description",
						Status:                    "Started",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", "Started", tm, atc, tm, 1).WillReturnResult(sqlMock.NewErrorResult(errors.New("RowsAffected failed")))
			},
			wantErr: true,
		},
		{
			name: "Not Found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: apiVersion,
					ToDo: &v1.ToDo{
						Id:                        1,
						Title:                     "new title",
						Description:               "new description",
						Status:                    "Started",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
						ActualTimeOfCompletion:    actualTimeOfCompletion,
						Reminder:                  reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", tm, atc, tm, 1).WillReturnResult(sqlMock.NewResult(1, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Update(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDoServiceServerDelete(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlMock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)

	type args struct {
		ctx context.Context
		req *v1.DeleteRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.DeleteResponse
		wantErr bool
	}{
		{name: "OK",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).WillReturnResult(sqlMock.NewResult(1, 1))
			},
			want: &v1.DeleteResponse{
				Api:     "v1",
				Deleted: 1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "DELETE failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).WillReturnError(errors.New("DELETE failed"))
			},
			wantErr: true,
		},
		{
			name: "RowsAffected failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).WillReturnResult(sqlMock.NewErrorResult(errors.New("RowsAffected failed")))
			},
			wantErr: true,
		},
		{
			name: "Not Found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: apiVersion,
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).WillReturnResult(sqlMock.NewResult(1, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Delete(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDoServiceServerReadAll(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlMock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	t1 := time.Now().In(time.UTC)
	tm1 := time.Date(2019, 9, 27, 8, 30, 45, 56478, time.UTC)
	reminder1, _ := ptypes.TimestampProto(t1)
	estimatedTimeOfCompletion1, _ := ptypes.TimestampProto(t1)
	actualTimeOfCompletion1, _ := ptypes.TimestampProto(tm1)
	t2 := time.Now().In(time.UTC)
	tm2 := time.Date(2019, 10, 26, 8, 30, 50, 23474, time.UTC)
	reminder2, _ := ptypes.TimestampProto(t2)
	estimatedTimeOfCompletion2, _ := ptypes.TimestampProto(t2)
	actualTimeOfCompletion2, _ := ptypes.TimestampProto(tm2)

	type args struct {
		ctx context.Context
		req *v1.ReadAllRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.ReadAllResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: apiVersion,
				},
			},
			mock: func() {
				rows := sqlMock.NewRows([]string{"ID", "Title", "Description", "Status", "EstimatedTimeOfCompletion", "ActualTimeOfCompletion", "Reminder"}).AddRow(1, "title 1", "description 1", "Completed", t1, tm1, t1).AddRow(2, "title 2", "description 2", "InProgress", t2, tm2, t2)
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WillReturnRows(rows)
			},
			want: &v1.ReadAllResponse{
				Api: "v1",
				ToDos: []*v1.ToDo{
					{
						Id:                        1,
						Title:                     "title 1",
						Status:                    "Completed",
						Description:               "description 1",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion1,
						ActualTimeOfCompletion:    actualTimeOfCompletion1,
						Reminder:                  reminder1,
					},
					{
						Id:                        2,
						Title:                     "title 2",
						Status:                    "InProgress",
						Description:               "description 2",
						EstimatedTimeOfCompletion: estimatedTimeOfCompletion2,
						ActualTimeOfCompletion:    actualTimeOfCompletion2,
						Reminder:                  reminder2,
					},
				},
			},
		},
		{
			name: "Empty",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: apiVersion,
				},
			},
			mock: func() {
				rows := sqlMock.NewRows([]string{"ID", "Title", "Status", "Description", "EstimatedTimeOfCompletion", "ActualTimeOfCompletion", "Reminder"})
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WillReturnRows(rows)
			},
			want: &v1.ReadAllResponse{
				Api:   "v1",
				ToDos: []*v1.ToDo{},
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: "v2",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.ReadAll(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoService.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.ReadAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
