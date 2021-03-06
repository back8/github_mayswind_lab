<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :back-link="$t('Back')"></f7-nav-left>
            <f7-nav-title :title="$t('User Profile')"></f7-nav-title>
            <f7-nav-right>
                <f7-link :class="{ 'disabled': inputIsNotChanged || inputIsInvalid || saving }" :text="$t('Save')" @click="save"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-card class="skeleton-text" v-if="loading">
            <f7-card-content class="no-safe-areas" :padding="false">
                <f7-list inline-labels>
                    <f7-list-input label="Password" placeholder="Your password"></f7-list-input>
                    <f7-list-input label="Confirmation Password" placeholder="Re-enter the password"></f7-list-input>
                    <f7-list-input label="E-mail" placeholder="Your email address"></f7-list-input>
                    <f7-list-input label="Nickname" placeholder="Your nickname"></f7-list-input>
                    <f7-list-item title="Default Currency" after="Currency"></f7-list-item>
                </f7-list>
            </f7-card-content>
        </f7-card>

        <f7-card v-else-if="!loading">
            <f7-card-content class="no-safe-areas" :padding="false">
                <f7-list form inline-labels>
                    <f7-list-input
                        type="password"
                        autocomplete="new-password"
                        clear-button
                        :label="$t('Password')"
                        :placeholder="$t('Your password')"
                        :value="newProfile.password"
                        @input="newProfile.password = $event.target.value"
                    ></f7-list-input>

                    <f7-list-input
                        type="password"
                        autocomplete="new-password"
                        clear-button
                        :label="$t('Confirmation Password')"
                        :placeholder="$t('Re-enter the password')"
                        :value="newProfile.confirmPassword"
                        @input="newProfile.confirmPassword = $event.target.value"
                    ></f7-list-input>

                    <f7-list-input
                        type="email"
                        autocomplete="email"
                        clear-button
                        :label="$t('E-mail')"
                        :placeholder="$t('Your email address')"
                        :value="newProfile.email"
                        @input="newProfile.email = $event.target.value"
                    ></f7-list-input>

                    <f7-list-input
                        type="text"
                        autocomplete="nickname"
                        clear-button
                        :label="$t('Nickname')"
                        :placeholder="$t('Your nickname')"
                        :value="newProfile.nickname"
                        @input="newProfile.nickname = $event.target.value"
                    ></f7-list-input>

                    <f7-list-item
                        :title="$t('Default Currency')"
                        smart-select :smart-select-params="{ openIn: 'popup', searchbar: true, searchbarPlaceholder: $t('Currency Name'), searchbarDisableText: $t('Cancel'), closeOnSelect: true, popupCloseLinkText: $t('Close'), scrollToSelectedItem: true }"
                    >
                        <select autocomplete="transaction-currency" v-model="newProfile.defaultCurrency">
                            <option v-for="currency in allCurrencies"
                                    :key="currency.code"
                                    :value="currency.code">{{ currency.displayName }}</option>
                        </select>
                    </f7-list-item>

                    <f7-list-item class="lab-list-item-error-info" v-if="inputIsInvalid" :footer="$t(inputInvalidProblemMessage)"></f7-list-item>
                </f7-list>
            </f7-card-content>
        </f7-card>

        <password-input-sheet :title="$t('Current Password')"
                              :hint="$t('Please enter your current password when modifying your password')"
                              :show.sync="showInputPasswordSheet"
                              :confirm-disabled="saving"
                              :cancel-disabled="saving"
                              v-model="currentPassword"
                              @password:confirm="save()">
        </password-input-sheet>
    </f7-page>
</template>

<script>
export default {
    data() {
        return {
            newProfile: {
                password: '',
                confirmPassword: '',
                email: '',
                nickname: '',
                defaultCurrency: ''
            },
            oldProfile: {
                email: '',
                nickname: '',
                defaultCurrency: ''
            },
            currentPassword: '',
            loading: true,
            saving: false,
            showInputPasswordSheet: false
        };
    },
    computed: {
        allCurrencies() {
            return this.$locale.getAllCurrencies();
        },
        inputIsNotChanged() {
            return !!this.inputIsNotChangedProblemMessage;
        },
        inputIsInvalid() {
            return !!this.inputInvalidProblemMessage;
        },
        inputIsNotChangedProblemMessage() {
            if (!this.newProfile.password && !this.newProfile.confirmPassword && !this.newProfile.email && !this.newProfile.nickname) {
                return 'Nothing has been modified';
            } else if (!this.newProfile.password && !this.newProfile.confirmPassword &&
                this.newProfile.email === this.oldProfile.email &&
                this.newProfile.nickname === this.oldProfile.nickname &&
                this.newProfile.defaultCurrency === this.oldProfile.defaultCurrency) {
                return 'Nothing has been modified';
            } else if (!this.newProfile.password && this.newProfile.confirmPassword) {
                return 'Password cannot be empty';
            } else if (this.newProfile.password && !this.newProfile.confirmPassword) {
                return 'Confirmation password cannot be empty';
            } else {
                return null;
            }
        },
        inputInvalidProblemMessage() {
            if (this.newProfile.password && this.newProfile.confirmPassword && this.newProfile.password !== this.newProfile.confirmPassword) {
                return 'Password and confirmation password do not match';
            } else if (!this.newProfile.email) {
                return 'Email address cannot be empty';
            } else if (!this.newProfile.nickname) {
                return 'Nickname cannot be empty';
            } else if (!this.newProfile.defaultCurrency) {
                return 'Default currency cannot be empty';
            } else {
                return null;
            }
        }
    },
    created() {
        const self = this;
        const router = self.$f7router;

        self.loading = true;

        self.$services.getProfile().then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                self.$toast('Unable to get user profile');
                router.back();
                return;
            }

            self.oldProfile.email = data.result.email;
            self.oldProfile.nickname = data.result.nickname;
            self.oldProfile.defaultCurrency = data.result.defaultCurrency;

            self.newProfile.email = self.oldProfile.email
            self.newProfile.nickname = self.oldProfile.nickname;
            self.newProfile.defaultCurrency = self.oldProfile.defaultCurrency;
            self.loading = false;
        }).catch(error => {
            self.$logger.error('failed to get user profile', error);

            if (error.response && error.response.data && error.response.data.errorMessage) {
                self.$toast({ error: error.response.data });
                router.back();
            } else if (!error.processed) {
                self.$toast('Unable to get user profile');
                router.back();
            }
        });
    },
    methods: {
        save() {
            const self = this;
            const router = self.$f7router;

            self.showInputPasswordSheet = false;

            let problemMessage = self.inputIsNotChangedProblemMessage || self.inputInvalidProblemMessage;

            if (problemMessage) {
                self.$alert(problemMessage);
                return;
            }

            if (self.newProfile.password && !self.currentPassword) {
                self.showInputPasswordSheet = true;
                return;
            }

            self.saving = true;
            self.$showLoading(() => self.saving);

            self.$services.updateProfile({
                password: self.newProfile.password,
                oldPassword: self.currentPassword,
                email: self.newProfile.email,
                nickname: self.newProfile.nickname,
                defaultCurrency: self.newProfile.defaultCurrency
            }).then(response => {
                self.saving = false;
                self.$hideLoading();
                self.currentPassword = '';

                const data = response.data;

                if (!data || !data.success || !data.result) {
                    self.$toast('Unable to update user profile');
                    return;
                }

                if (self.$utilities.isString(data.result.newToken)) {
                    self.$user.updateToken(data.result.newToken);
                }

                if (self.$utilities.isObject(data.result.user)) {
                    self.$user.updateUserInfo(data.result.user);
                }

                self.$toast('Your profile has been successfully updated');
                router.back();
            }).catch(error => {
                self.$logger.error('failed to save user profile', error);

                self.saving = false;
                self.$hideLoading();
                self.currentPassword = '';

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    self.$toast({ error: error.response.data });
                } else if (!error.processed) {
                    self.$toast('Unable to update user profile');
                }
            });
        }
    }
};
</script>
