package v1

import (
	"context"
	"database/sql"
	"time"

	v1 "github.com/basebandit/go-grpc/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	//apiVersion is version of API as provided by server
	apiVersion = "v1"
)

//todoServiceServer is implementation of v1.ToDoServiceServer proto interface
type todoServiceServer struct {
	db *sql.DB
}

//NewToDoServiceServer creates ToDo service server
func NewToDoServiceServer(db *sql.DB) v1.ToDoServiceServer {
	return &todoServiceServer{db: db}
}

//checkAPI checks if the API version requested by client is supported by server
func (s *todoServiceServer) checkAPI(api string) error {
	//API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented, "unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

//connect returns SQL database connection from the pool
func (s *todoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to connect to database -> %s", err.Error())
	}
	return c, nil
}

//Create creates a new todo entity
func (s *todoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	//check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//get SQL connection from the connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "reminder field has invalid format -> %s", err.Error())
	}

	estimatedTimeOfCompletion, err := ptypes.Timestamp(req.ToDo.EstimatedTimeOfCompletion)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "estimatedTimeOfCompletion has invalid format -> %s", err.Error())
	}

	//insert todo entity data
	res, err := c.ExecContext(ctx, "INSERT INTO ToDo(`Title`,`Description`,`Status`,`EstimatedTimeOfCompletion`,`ActualTimeOfCompletion`,`Reminder`) VALUES (?,?,?,?,?,?)", req.ToDo.Title, req.ToDo.Description, req.ToDo.Status, estimatedTimeOfCompletion, estimatedTimeOfCompletion, reminder)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to insert into ToDo -> %s", err.Error())
	}

	//get ID of created todo entity
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve id for created ToDo -> %s", err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

//Read reads todo entity
func (s *todoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	//check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//get SQL connection from  the connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	//query todo entity by ID
	rows, err := c.QueryContext(ctx, "SELECT `ID`,`Title`,`Description`,`Status`,`EstimatedTimeOfCompletion`,`ActualTimeOfCompletion` ,`Reminder` FROM ToDo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to select from ToDo -> %s", err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to retrieve data from ToDo -> %s", err.Error())
		}

		return nil, status.Errorf(codes.NotFound, "ToDo with ID='%d' is not found", req.Id)
	}

	//get todo entity
	var td v1.ToDo
	var estimatedTimeOfCompletion time.Time
	var actualTimeOfCompletion time.Time
	var reminder time.Time

	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &td.Status, &estimatedTimeOfCompletion, &actualTimeOfCompletion, &reminder); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve field values from ToDo row -> %s", err.Error())
	}
	td.EstimatedTimeOfCompletion, err = ptypes.TimestampProto(estimatedTimeOfCompletion)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "estimatedTimeOfCompletion field has invalid format -> %s", err.Error())
	}

	td.ActualTimeOfCompletion, err = ptypes.TimestampProto(actualTimeOfCompletion)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "actualTimeOfCompletion field has invalid format -> %s", err.Error())
	}

	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "reminder field has invalid format -> %s", err.Error())
	}

	if rows.Next() {
		return nil, status.Errorf(codes.Unknown, "found multiple ToDo rows with ID='%d'", req.Id)
	}

	return &v1.ReadResponse{
		Api:  apiVersion,
		ToDo: &td,
	}, nil
}

//Update updates a todo entity
func (s *todoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	//check if the API version requested by the client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//get SQL connection from the connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	estimatedTimeOfCompletion, err := ptypes.Timestamp(req.ToDo.EstimatedTimeOfCompletion)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "estimatedTimeOfCompletion field has invalid format -> %s", err.Error())
	}

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "reminder field has invalid format -> %s", err.Error())
	}

	var actualTimeOfCompletion time.Time

	if req.ToDo.Status == "Completed" {
		actualTimeOfCompletion = time.Date(2020, 2, 27, 3, 15, 45, 34567, time.UTC)
	} else {
		actualTimeOfCompletion, err = ptypes.Timestamp(req.ToDo.ActualTimeOfCompletion)
	}
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "actualTimeOfCompletion field has invalid format -> %s", err.Error())
	}

	//update todo entity
	res, err := c.ExecContext(ctx, "UPDATE ToDo SET `Title`=?, `Description`=?, `Status`=?, `EstimatedTimeOfCompletion`=?, `ActualTimeOfCompletion`=?,`Reminder`=? WHERE `ID`=?", req.ToDo.Title, req.ToDo.Description, req.ToDo.Status, estimatedTimeOfCompletion, actualTimeOfCompletion, reminder, req.ToDo.Id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to update ToDo -> %s", err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve rows affected value -> %s", err.Error())
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "ToDo with ID='%d' is not found", req.ToDo.Id)
	}
	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

//Delete deleted a todo entity
func (s *todoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	//check if the API version requested by the client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//get SQL connection from the connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	//delete todo entity
	res, err := c.ExecContext(ctx, "DELETE FROM ToDo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to delete ToDo -> %s", err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve rows affected value -> %s", err.Error())
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "ToDo with ID='%d' is not found", req.Id)
	}
	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

//Read all todo entities
func (s *todoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	//check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//get SQL connection from the connection pool
	c, err := s.connect(ctx)

	if err != nil {
		return nil, err
	}
	defer c.Close()

	//get todo entity list
	rows, err := c.QueryContext(ctx, "SELECT `ID`,`Title`,`Description`,`Status`,`EstimatedTimeOfCompletion`,`ActualTimeOfCompletion`,`Reminder` FROM ToDo")
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to select from ToDo -> %s", err.Error())
	}
	defer rows.Close()

	var estimatedTimeOfCompletion time.Time
	var actualTimeOfCompletion time.Time
	var reminder time.Time

	list := []*v1.ToDo{}
	for rows.Next() {
		td := new(v1.ToDo) //pointer to an empty todo struct initialized to default zero values of its respective fields
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &td.Status, &estimatedTimeOfCompletion, &actualTimeOfCompletion, &reminder); err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to retrieve field values from ToDo row -> %s", err.Error())
		}

		td.EstimatedTimeOfCompletion, err = ptypes.TimestampProto(estimatedTimeOfCompletion)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "estimatedTimeOfCompletion field has invalid format -> %s", err.Error())
		}

		td.ActualTimeOfCompletion, err = ptypes.TimestampProto(actualTimeOfCompletion)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "actualTimeOfCompletion field has invalid format -> %s", err.Error())
		}

		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "reminder field has invalid format -> %s", err.Error())
		}

		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve data from ToDo-> %s", err.Error())
	}

	return &v1.ReadAllResponse{
		Api:   apiVersion,
		ToDos: list,
	}, nil
}
