package errors

import (
	"errors"
	"fmt"

	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/tuple"
	openfgapb "go.buf.build/openfga/go/openfga/api/openfga/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const InternalServerErrorMsg = "Internal Server Error"

var (
	// AuthorizationModelResolutionTooComplex is used to avoid stack overflows
	AuthorizationModelResolutionTooComplex = status.Error(codes.Code(openfgapb.ErrorCode_authorization_model_resolution_too_complex), "Authorization Model resolution required too many rewrite rules to be resolved. Check your authorization model for infinite recursion or too much nesting")
	InvalidWriteInput                      = status.Error(codes.Code(openfgapb.ErrorCode_invalid_write_input), "Invalid input. Make sure you provide at least one write, or at least one delete")
	InvalidContinuationToken               = status.Error(codes.Code(openfgapb.ErrorCode_invalid_continuation_token), "Invalid continuation token")
	InvalidTupleSet                        = status.Error(codes.Code(openfgapb.ErrorCode_invalid_tuple_set), "Invalid TupleSet. Make sure you provide a type, and either an objectId or a userSet")
	InvalidCheckInput                      = status.Error(codes.Code(openfgapb.ErrorCode_invalid_check_input), "Invalid input. Make sure you provide a user, object and relation")
	InvalidExpandInput                     = status.Error(codes.Code(openfgapb.ErrorCode_invalid_expand_input), "Invalid input. Make sure you provide an object and a relation")
	UnsupportedUserSet                     = status.Error(codes.Code(openfgapb.ErrorCode_unsupported_user_set), "Userset is not supported (right now)")
	StoreIDNotFound                        = status.Error(codes.Code(openfgapb.NotFoundErrorCode_store_id_not_found), "Store ID not found")
	MismatchObjectType                     = status.Error(codes.Code(openfgapb.ErrorCode_query_string_type_continuation_token_mismatch), "The type in the querystring and the continuation token don't match")
	RequestCancelled                       = status.Error(codes.Code(openfgapb.InternalErrorCode_cancelled), "Request Cancelled")
)

type InternalError struct {
	public   error
	internal error
}

func (e InternalError) Error() string {
	return e.public.Error()
}

func (e InternalError) Is(target error) bool {
	return target.Error() == e.Error()
}

func (e InternalError) InternalError() string {
	return e.internal.Error()
}

func (e InternalError) Internal() error {
	return e.internal
}

func NewInternalError(public string, internal error) InternalError {
	if public == "" {
		public = InternalServerErrorMsg
	}

	return InternalError{
		public:   status.Error(codes.Code(openfgapb.InternalErrorCode_internal_error), public),
		internal: internal,
	}
}

func ValidationError(cause error) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_validation_error), cause.Error())
}

func AssertionsNotForAuthorizationModelFound(modelID string) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_authorization_model_assertions_not_found), fmt.Sprintf("No assertions found for authorization model '%s'", modelID))
}

func AuthorizationModelNotFound(modelID string) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_authorization_model_not_found), fmt.Sprintf("Authorization Model '%s' not found", modelID))
}

func LatestAuthorizationModelNotFound(store string) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_latest_authorization_model_not_found), fmt.Sprintf("No authorization models found for store '%s'", store))
}

func TypeNotFound(objectType string) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_type_not_found), fmt.Sprintf("type '%s' not found", objectType))
}

func RelationNotFound(relation string, objectType string, tk *openfgapb.TupleKey) error {

	msg := fmt.Sprintf("relation '%s#%s' not found", objectType, relation)
	if tk != nil {
		msg += fmt.Sprintf(" for tuple '%s'", tuple.TupleKeyToString(tk))
	}

	return status.Error(codes.Code(openfgapb.ErrorCode_relation_not_found), msg)
}

func ExceededEntityLimit(entity string, limit int) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_exceeded_entity_limit),
		fmt.Sprintf("The number of %s exceeds the allowed limit of %d", entity, limit))
}

func InvalidTuple(reason string, tuple *openfgapb.TupleKey) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_invalid_tuple), fmt.Sprintf("Invalid tuple '%s'. Reason: %s", tuple.String(), reason))
}

func DuplicateContextualTuple(tk *openfgapb.TupleKey) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_duplicate_contextual_tuple), fmt.Sprintf("Duplicate contextual tuple in request: %s.", tk.String()))
}

// InvalidObjectFormat is used when an object does not have a type and id part
func InvalidObjectFormat(tuple *openfgapb.TupleKey) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_invalid_object_format), fmt.Sprintf("Invalid object format for tuple '%s'", tuple.String()))
}

func DuplicateTupleInWrite(tk *openfgapb.TupleKey) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_cannot_allow_duplicate_tuples_in_one_request), fmt.Sprintf("duplicate tuple in write: user: '%s', relation: '%s', object: '%s'", tk.GetUser(), tk.GetRelation(), tk.GetObject()))
}

func WriteToIndirectRelationError(reason string, tk *openfgapb.TupleKey) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_invalid_tuple), fmt.Sprintf("Invalid tuple '%s'. Reason: %s", tk.String(), reason))
}

func WriteFailedDueToInvalidInput(err error) error {
	if err != nil {
		return status.Error(codes.Code(openfgapb.ErrorCode_write_failed_due_to_invalid_input), err.Error())
	}
	return status.Error(codes.Code(openfgapb.ErrorCode_write_failed_due_to_invalid_input), "Write failed due to invalid input")
}

func InvalidAuthorizationModelInput(err error) error {
	return status.Error(codes.Code(openfgapb.ErrorCode_invalid_authorization_model), err.Error())
}

// HandleError is used to hide internal errors from users. Use `public` to return an error message to the user.
func HandleError(public string, err error) error {
	if errors.Is(err, storage.ErrInvalidContinuationToken) {
		return InvalidContinuationToken
	} else if errors.Is(err, storage.ErrMismatchObjectType) {
		return MismatchObjectType
	} else if errors.Is(err, storage.ErrCancelled) {
		return RequestCancelled
	}
	return NewInternalError(public, err)
}

// HandleTupleValidateError provide common routines for handling tuples validation error
func HandleTupleValidateError(err error) error {
	switch t := err.(type) {
	case *tuple.InvalidTupleError:
		return InvalidTuple(t.Cause.Error(), t.TupleKey)
	case *tuple.InvalidObjectFormatError:
		return InvalidObjectFormat(t.TupleKey)
	case *tuple.TypeNotFoundError:
		return TypeNotFound(t.TypeName)
	case *tuple.RelationNotFoundError:
		return RelationNotFound(t.Relation, t.TypeName, t.TupleKey)
	case *tuple.IndirectWriteError:
		return WriteToIndirectRelationError(t.Reason, t.TupleKey)
	}

	return HandleError("", err)
}
