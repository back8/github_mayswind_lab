package api

import (
	"sort"

	"github.com/mayswind/lab/pkg/core"
	"github.com/mayswind/lab/pkg/errs"
	"github.com/mayswind/lab/pkg/log"
	"github.com/mayswind/lab/pkg/models"
	"github.com/mayswind/lab/pkg/services"
	"github.com/mayswind/lab/pkg/utils"
)

// TransactionsApi represents transaction api
type TransactionsApi struct {
	transactions    *services.TransactionService
	transactionTags *services.TransactionTagService
}

// Initialize a transaction api singleton instance
var (
	Transactions = &TransactionsApi{
		transactions:    services.Transactions,
		transactionTags: services.TransactionTags,
	}
)

// TransactionListHandler returns transaction list of current user
func (a *TransactionsApi) TransactionListHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionListReq models.TransactionListByMaxTimeRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	transactions, err := a.transactions.GetTransactionsByMaxTime(uid, transactionListReq.MaxTime, transactionListReq.MinTime, transactionListReq.Type, transactionListReq.CategoryId, transactionListReq.AccountId, transactionListReq.Keyword, transactionListReq.Count+1)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionListHandler] failed to get transactions earlier than \"%d\" for user \"uid:%d\", because %s", transactionListReq.MaxTime, uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	finalCount := transactionListReq.Count

	if len(transactions) < finalCount {
		finalCount = len(transactions)
	}

	hasMore := false

	if finalCount < len(transactions) {
		hasMore = true
	}

	transactionIds := make([]int64, finalCount)

	for i := 0; i < finalCount; i++ {
		transactionIds[i] = transactions[i].TransactionId
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(uid, transactionIds)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionListHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	transactionResps := &models.TransactionInfoPageWrapperResponse{}
	transactionResps.Items = make(models.TransactionInfoResponseSlice, finalCount)

	for i := 0; i < finalCount; i++ {
		transactionTagIds := allTransactionTagIds[transactions[i].TransactionId]
		transactionResps.Items[i] = transactions[i].ToTransactionInfoResponse(transactionTagIds)
	}

	sort.Sort(transactionResps.Items)

	if hasMore {
		transactionResps.NextTimeSequenceId = &transactions[finalCount].TransactionTime
	}

	return transactionResps, nil
}

// TransactionMonthListHandler returns transaction list of current user by month
func (a *TransactionsApi) TransactionMonthListHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionListReq models.TransactionListInMonthByPageRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionMonthListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	transactions, err := a.transactions.GetTransactionsInMonthByPage(uid, transactionListReq.Year, transactionListReq.Month, transactionListReq.Type, transactionListReq.CategoryId, transactionListReq.AccountId, transactionListReq.Keyword, transactionListReq.Page, transactionListReq.Count)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionMonthListHandler] failed to get transactions in month \"%d-%d\" for user \"uid:%d\", because %s", transactionListReq.Year, transactionListReq.Month, uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	transactionIds := make([]int64, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionIds[i] = transactions[i].TransactionId
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(uid, transactionIds)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionMonthListHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	transactionResps := make([]*models.TransactionInfoResponse, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionTagIds := allTransactionTagIds[transactions[i].TransactionId]
		transactionResps[i] = transactions[i].ToTransactionInfoResponse(transactionTagIds)
	}

	return transactionResps, nil
}

// TransactionGetHandler returns one specific transaction of current user
func (a *TransactionsApi) TransactionGetHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionGetReq models.TransactionGetRequest
	err := c.ShouldBindQuery(&transactionGetReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	transaction, err := a.transactions.GetTransactionByTransactionId(uid, transactionGetReq.Id)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionGetHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionGetReq.Id, uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction = a.transactions.GetRelatedTransferTransaction(transaction, transaction.RelatedId)
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(uid, []int64{transaction.TransactionId})

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionGetHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	transactionTagIds := allTransactionTagIds[transaction.TransactionId]
	transactionResp := transaction.ToTransactionInfoResponse(transactionTagIds)

	return transactionResp, nil
}

// TransactionCreateHandler saves a new transaction by request parameters for current user
func (a *TransactionsApi) TransactionCreateHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionCreateReq models.TransactionCreateRequest
	err := c.ShouldBindJSON(&transactionCreateReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionCreateReq.TagIds)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if transactionCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] transaction type is invalid")
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionCreateReq.CategoryId > 0 {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAccountId != 0 {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination account cannot be set")
		return nil, errs.ErrTransactionDestinationAccountCannotBeSet
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.SourceAccountId == transactionCreateReq.DestinationAccountId {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] transfer transaction source account must not be destination account")
		return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAmount != 0 {
		log.WarnfWithRequestId(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination amount cannot be set")
		return nil, errs.ErrTransactionDestinationAmountCannotBeSet
	}

	uid := c.GetCurrentUid()
	transaction := a.createNewTransactionModel(uid, &transactionCreateReq)

	err = a.transactions.CreateTransaction(transaction, tagIds)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionCreateHandler] failed to create transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.InfofWithRequestId(c, "[transactions.TransactionCreateHandler] user \"uid:%d\" has created a new transaction \"id:%d\" successfully", uid, transaction.TransactionId)

	transactionResp := transaction.ToTransactionInfoResponse(nil)

	return transactionResp, nil
}

// TransactionModifyHandler saves an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionModifyHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionModifyReq models.TransactionModifyRequest
	err := c.ShouldBindJSON(&transactionModifyReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionModifyReq.TagIds)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionModifyHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	uid := c.GetCurrentUid()
	transaction, err := a.transactions.GetTransactionByTransactionId(uid, transactionModifyReq.Id)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionModifyHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.ErrorfWithRequestId(c, "[transactions.TransactionModifyHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transactionModifyReq.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(uid, []int64{transaction.TransactionId})

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionModifyHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	transactionTagIds := allTransactionTagIds[transaction.TransactionId]
	addTransactionTagIds := utils.Int64SliceMinus(tagIds, transactionTagIds)
	removeTransactionTagIds := utils.Int64SliceMinus(transactionTagIds, tagIds)

	newTransaction := &models.Transaction{
		TransactionId:   transaction.TransactionId,
		Uid:             uid,
		CategoryId:      transactionModifyReq.CategoryId,
		TransactionTime: utils.GetMinTransactionTimeFromUnixTime(transactionModifyReq.Time),
		AccountId:       transactionModifyReq.SourceAccountId,
		Amount:          transactionModifyReq.SourceAmount,
		Comment:         transactionModifyReq.Comment,
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		newTransaction.RelatedAccountId = transactionModifyReq.DestinationAccountId
		newTransaction.RelatedAccountAmount = transactionModifyReq.DestinationAmount
	}

	if newTransaction.CategoryId == transaction.CategoryId &&
		utils.GetUnixTimeFromTransactionTime(newTransaction.TransactionTime) == utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) &&
		newTransaction.AccountId == transaction.AccountId &&
		newTransaction.Amount == transaction.Amount &&
		(transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountId == transaction.RelatedAccountId) &&
		(transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountAmount == transaction.RelatedAccountAmount) &&
		newTransaction.Comment == transaction.Comment &&
		len(addTransactionTagIds) < 1 &&
		len(removeTransactionTagIds) < 1 {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.transactions.ModifyTransaction(newTransaction, addTransactionTagIds, removeTransactionTagIds)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionModifyHandler] failed to update transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.InfofWithRequestId(c, "[transactions.TransactionModifyHandler] user \"uid:%d\" has updated transaction \"id:%d\" successfully", uid, transactionModifyReq.Id)

	newTransaction.Type = transaction.Type
	newTransactionResp := newTransaction.ToTransactionInfoResponse(tagIds)

	return newTransactionResp, nil
}

// TransactionDeleteHandler deletes an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionDeleteHandler(c *core.Context) (interface{}, *errs.Error) {
	var transactionDeleteReq models.TransactionDeleteRequest
	err := c.ShouldBindJSON(&transactionDeleteReq)

	if err != nil {
		log.WarnfWithRequestId(c, "[transactions.TransactionDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.transactions.DeleteTransaction(uid, transactionDeleteReq.Id)

	if err != nil {
		log.ErrorfWithRequestId(c, "[transactions.TransactionDeleteHandler] failed to delete transaction \"id:%d\" for user \"uid:%d\", because %s", transactionDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.InfofWithRequestId(c, "[transactions.TransactionDeleteHandler] user \"uid:%d\" has deleted transaction \"id:%d\"", uid, transactionDeleteReq.Id)
	return true, nil
}

func (a *TransactionsApi) createNewTransactionModel(uid int64, transactionCreateReq *models.TransactionCreateRequest) *models.Transaction {
	var transactionDbType models.TransactionDbType

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE {
		transactionDbType = models.TRANSACTION_DB_TYPE_MODIFY_BALANCE
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_EXPENSE {
		transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_INCOME {
		transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	}

	transaction := &models.Transaction{
		Uid:             uid,
		Type:            transactionDbType,
		CategoryId:      transactionCreateReq.CategoryId,
		TransactionTime: utils.GetMinTransactionTimeFromUnixTime(transactionCreateReq.Time),
		AccountId:       transactionCreateReq.SourceAccountId,
		Amount:          transactionCreateReq.SourceAmount,
		Comment:         transactionCreateReq.Comment,
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		transaction.RelatedAccountId = transactionCreateReq.DestinationAccountId
		transaction.RelatedAccountAmount = transactionCreateReq.DestinationAmount
	}

	return transaction
}
