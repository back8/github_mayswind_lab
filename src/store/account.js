import services from '../lib/services.js';
import logger from '../lib/logger.js';

import {
    LOAD_ACCOUNT_LIST,
    ADD_ACCOUNT_TO_ACCOUNT_LIST,
    SAVE_ACCOUNT_IN_ACCOUNT_LIST,
    CHANGE_ACCOUNT_DISPLAY_ORDER_IN_ACCOUNT_LIST,
    UPDATE_ACCOUNT_VISIBILITY_IN_ACCOUNT_LIST,
    REMOVE_ACCOUNT_FROM_ACCOUNT_LIST,
    UPDATE_ACCOUNT_LIST_INVALID_STATE
} from './mutations.js';

function loadAllAccounts(context, { force }) {
    if (!force && !context.state.accountListStateInvalid) {
        return new Promise((resolve) => {
            resolve(context.state.allAccounts);
        });
    }

    return new Promise((resolve, reject) => {
        services.getAllAccounts({
            visibleOnly: false
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                reject({ message: 'Unable to get account list' });
                return;
            }

            context.commit(LOAD_ACCOUNT_LIST, data.result);
            context.commit(UPDATE_ACCOUNT_LIST_INVALID_STATE, false);

            resolve(data.result);
        }).catch(error => {
            if (force) {
                logger.error('failed to force load account list', error);
            } else {
                logger.error('failed to load account list', error);
            }

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                reject({ message: 'Unable to get account list' });
            }
        });
    });
}

function getAccount(context, { accountId }) {
    return new Promise((resolve, reject) => {
        services.getAccount({
            id: accountId
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                reject({ message: 'Unable to get account' });
                return;
            }

            resolve(data.result);
        }).catch(error => {
            logger.error('failed to load account info', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                reject({ message: 'Unable to get account' });
            }
        });
    });
}

function saveAccount(context, { account }) {
    return new Promise((resolve, reject) => {
        let promise = null;

        if (!account.id) {
            promise = services.addAccount(account);
        } else {
            promise = services.modifyAccount(account);
        }

        promise.then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                if (!account.id) {
                    reject({ message: 'Unable to add account' });
                } else {
                    reject({ message: 'Unable to save account' });
                }
                return;
            }

            if (!account.id) {
                context.commit(ADD_ACCOUNT_TO_ACCOUNT_LIST, data.result);
            } else {
                context.commit(SAVE_ACCOUNT_IN_ACCOUNT_LIST, data.result);
            }

            resolve(data.result);
        }).catch(error => {
            logger.error('failed to save account', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                if (!account.id) {
                    reject({ message: 'Unable to add account' });
                } else {
                    reject({ message: 'Unable to save account' });
                }
            }
        });
    });
}

function changeAccountDisplayOrder(context, { accountId, from, to }) {
    const account = context.state.allAccountsMap[accountId];

    return new Promise((resolve, reject) => {
        if (!account ||
            !context.state.allCategorizedAccounts[account.category] ||
            !context.state.allCategorizedAccounts[account.category].accounts ||
            !context.state.allCategorizedAccounts[account.category].accounts[to]) {
            reject({ message: 'Unable to move account' });
            return;
        }

        context.commit(UPDATE_ACCOUNT_LIST_INVALID_STATE, true);
        context.commit(CHANGE_ACCOUNT_DISPLAY_ORDER_IN_ACCOUNT_LIST, {
            account: account,
            from: from,
            to: to
        });

        resolve();
    });
}

function updateAccountDisplayOrders(context) {
    const newDisplayOrders = [];

    for (let category in context.state.allCategorizedAccounts) {
        if (!Object.prototype.hasOwnProperty.call(context.state.allCategorizedAccounts, category)) {
            continue;
        }

        const accountList = context.state.allCategorizedAccounts[category].accounts;

        for (let i = 0; i < accountList.length; i++) {
            newDisplayOrders.push({
                id: accountList[i].id,
                displayOrder: i + 1
            });
        }
    }

    return new Promise((resolve, reject) => {
        services.moveAccount({
            newDisplayOrders: newDisplayOrders
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                reject({ message: 'Unable to move account' });
                return;
            }

            context.commit(UPDATE_ACCOUNT_LIST_INVALID_STATE, false);

            resolve(data.result);
        }).catch(error => {
            logger.error('failed to save accounts display order', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                reject({ message: 'Unable to move account' });
            }
        });
    });
}

function hideAccount(context, { account, hidden }) {
    return new Promise((resolve, reject) => {
        services.hideAccount({
            id: account.id,
            hidden: hidden
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                if (hidden) {
                    reject({ message: 'Unable to hide this account' });
                } else {
                    reject({ message: 'Unable to unhide this account' });
                }

                return;
            }

            context.commit(UPDATE_ACCOUNT_VISIBILITY_IN_ACCOUNT_LIST, {
                account: account,
                hidden: hidden
            });

            resolve(data.result);
        }).catch(error => {
            logger.error('failed to change account visibility', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                if (hidden) {
                    reject({ message: 'Unable to hide this account' });
                } else {
                    reject({ message: 'Unable to unhide this account' });
                }
            }
        });
    });
}

function deleteAccount(context, { account, beforeResolve }) {
    return new Promise((resolve, reject) => {
        services.deleteAccount({
            id: account.id
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                reject({ message: 'Unable to delete this account' });
                return;
            }

            if (beforeResolve) {
                beforeResolve(() => {
                    context.commit(REMOVE_ACCOUNT_FROM_ACCOUNT_LIST, account);
                });
            } else {
                context.commit(REMOVE_ACCOUNT_FROM_ACCOUNT_LIST, account);
            }

            resolve(data.result);
        }).catch(error => {
            logger.error('failed to delete account', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                reject({ error: error.response.data });
            } else if (!error.processed) {
                reject({ message: 'Unable to delete this account' });
            }
        });
    });
}

function allAvailableAccountsCount(state) {
    let allAccountCount = 0;

    for (let category in state.allCategorizedAccounts) {
        if (!Object.prototype.hasOwnProperty.call(state.allCategorizedAccounts, category)) {
            continue;
        }

        allAccountCount += state.allCategorizedAccounts[category].accounts.length;
    }

    return allAccountCount;
}

function allVisibleAccountsCount(state) {
    let shownAccountCount = 0;

    for (let category in state.allCategorizedAccounts) {
        if (!Object.prototype.hasOwnProperty.call(state.allCategorizedAccounts, category)) {
            continue;
        }

        const accountList = state.allCategorizedAccounts[category].accounts;

        for (let i = 0; i < accountList.length; i++) {
            if (!accountList[i].hidden) {
                shownAccountCount++;
            }
        }
    }

    return shownAccountCount;
}

export default {
    loadAllAccounts,
    getAccount,
    saveAccount,
    changeAccountDisplayOrder,
    updateAccountDisplayOrders,
    hideAccount,
    deleteAccount,
    allAvailableAccountsCount,
    allVisibleAccountsCount,
}
