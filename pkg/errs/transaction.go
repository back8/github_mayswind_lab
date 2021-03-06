package errs

import "net/http"

// Error codes related to transaction
var (
	ErrTransactionIdInvalid                                = NewNormalError(NormalSubcategoryTransaction, 0, http.StatusBadRequest, "transaction id is invalid")
	ErrTransactionNotFound                                 = NewNormalError(NormalSubcategoryTransaction, 1, http.StatusBadRequest, "transaction not found")
	ErrTransactionTypeInvalid                              = NewNormalError(NormalSubcategoryTransaction, 2, http.StatusBadRequest, "transaction type is invalid")
	ErrTransactionSourceAndDestinationIdCannotBeEqual      = NewNormalError(NormalSubcategoryTransaction, 3, http.StatusBadRequest, "transaction source and destination account id cannot be equal")
	ErrTransactionSourceAndDestinationAmountNotEqual       = NewNormalError(NormalSubcategoryTransaction, 4, http.StatusBadRequest, "transaction source and destination amount not equal")
	ErrTransactionDestinationAccountCannotBeSet            = NewNormalError(NormalSubcategoryTransaction, 5, http.StatusBadRequest, "transaction destination account cannot be set")
	ErrTransactionDestinationAmountCannotBeSet             = NewNormalError(NormalSubcategoryTransaction, 6, http.StatusBadRequest, "transaction destination amount cannot be set")
	ErrTooMuchTransactionInOneSecond                       = NewNormalError(NormalSubcategoryTransaction, 7, http.StatusBadRequest, "too much transaction in one second")
	ErrBalanceModificationTransactionCannotSetCategory     = NewNormalError(NormalSubcategoryTransaction, 8, http.StatusBadRequest, "balance modification transaction cannot set category")
	ErrBalanceModificationTransactionCannotChangeAccountId = NewNormalError(NormalSubcategoryTransaction, 9, http.StatusBadRequest, "balance modification transaction cannot change account id")
	ErrBalanceModificationTransactionCannotAddWhenNotEmpty = NewNormalError(NormalSubcategoryTransaction, 10, http.StatusBadRequest, "balance modification transaction cannot add when other transaction exists")
	ErrCannotAddTransactionToHiddenAccount                 = NewNormalError(NormalSubcategoryTransaction, 11, http.StatusBadRequest, "cannot add transaction to hidden account")
	ErrCannotModifyTransactionInHiddenAccount              = NewNormalError(NormalSubcategoryTransaction, 12, http.StatusBadRequest, "cannot modify transaction of hidden account")
	ErrCannotDeleteTransactionInHiddenAccount              = NewNormalError(NormalSubcategoryTransaction, 13, http.StatusBadRequest, "cannot delete transaction in hidden account")
)
