package service

var (
	// user api errors

	ErrInvalidEmail             = &Error{Code: -1000, Message: "invalid email"}
	ErrInvalidPassword          = &Error{Code: -1001, Message: "invalid password"}
	ErrInvalidUserType          = &Error{Code: -1002, Message: "invalid user type"}
	ErrPasswordMismatch         = &Error{Code: -1003, Message: "password and confirm password must match"}
	ErrEmailNotExists           = &Error{Code: -1004, Message: "email doesn't exist"}
	ErrEmailAlreadyExists       = &Error{Code: -1005, Message: "email already exists"}
	ErrInvalidCredentials       = &Error{Code: -1006, Message: "invalid credentials"}
	ErrInvalidVerificationToken = &Error{Code: -1007, Message: "invalid verification token"}
	ErrInactiveAccount          = &Error{Code: -1008, Message: "your account is inactive"}
	ErrMissingPubKey            = &Error{Code: -1009, Message: "public key is required for lender user"}
	ErrInvalidUser              = &Error{Code: -1010, Message: "invalid user"}
	ErrEmailIsNotVerified       = &Error{Code: -1016, Message: "email is not verified"}
	ErrDeviceAlreadyExists      = &Error{Code: -1017, Message: "device id already exists"}
	ErrInvalidDevice            = &Error{Code: -1018, Message: "invalid device id"}
	AddressNotFound             = &Error{Code: -1019, Message: "Most of the address is busy, please try later."}
	MissTokenAddress            = &Error{Code: -1020, Message: "Token address is required"}

	HistoryNotFound    = &Error{Code: -1021, Message: "Invalid history"}
	TokenNotFound      = &Error{Code: -1022, Message: "Token is not found"}
	TokenAlreadyExists = &Error{Code: -1023, Message: "The token already exists"}
	InsufficientFunds  = &Error{Code: -1024, Message: "InsufficientFunds"}

	EmailAlreadyVified = &Error{Code: -1025, Message: "The email already verified"}
	TokenExpired       = &Error{Code: -1026, Message: "The token is expired"}
	MemoRequired       = &Error{Code: -1027, Message: "The Memo is required"}

	WithdrawalExists = &Error{Code: -1028, Message: "Please wait until the previous record is processed."}

	// exchange api errors

	ErrInvalidOrderType    = &Error{Code: -3000, Message: "invalid order type"}
	ErrInvalidOrderSide    = &Error{Code: -3001, Message: "invalid order side"}
	ErrInvalidSymbol       = &Error{Code: -3002, Message: "invalid symbol"}
	ErrInvalidOrderStatus  = &Error{Code: -3003, Message: "invalid order status"}
	ErrInvalidOrder        = &Error{Code: -3004, Message: "invalid order"}
	ErrInsufficientBalance = &Error{Code: -3005, Message: "insufficient balance"}

	// reserve api errors

	ErrInvalidReserve            = &Error{Code: -6000, Message: "invalid reserve"}
	ErrInvalidReserveType        = &Error{Code: -6001, Message: "invalid reserve type"}
	ErrInvalidReserveBuyingAsset = &Error{Code: -6002, Message: "invalid reserve buying asset"}

	// storage api errors
	ErrFileInvalidRequest = &Error{Code: -8000, Message: "invalid request"}
	ErrFileInvalid        = &Error{Code: -8001, Message: "file invalid"}
	ErrSizeLimit          = &Error{Code: -8002, Message: "invalid file size"}
	ErrFileType           = &Error{Code: -8003, Message: "invalid file type"}
	ErrSubmitFileFail     = &Error{Code: -8004, Message: "submit file failed"}

	// general api errors

	ErrInvalidArgument     = &Error{Code: -9000, Message: "invalid argument"}
	ErrInternalServerError = &Error{Code: -9001, Message: "internal server error"}
	ErrInvalidLimit        = &Error{Code: -9002, Message: "invalid limit"}
	ErrInvalidPage         = &Error{Code: -9003, Message: "invalid page"}
	ErrInvalidAmount       = &Error{Code: -9004, Message: "invalid amount"}
	ErrPermissionDenied    = &Error{Code: -9005, Message: "permission denied"}

	// faucet

	ErrAddressAlreadyGiven = &Error{Code: -10000, Message: "Your address is given constant. Please wait at least 12 hours for next request."}
	ErrInvalidURL          = &Error{Code: -10001, Message: "Invalid url."}
	ErrInvalidContent      = &Error{Code: -10002, Message: "Invalid content."}
	ErrInvalidFaucetAmount = &Error{Code: -10003, Message: "Your requested amount is too large."}

	// game
	ErrPendingTransaction         = &Error{Code: -70000, Message: "Player has pending transaction."}
	ErrPlayerNotFound             = &Error{Code: -70001, Message: "Player not found."}
	ErrNumberLargerThanRemaining  = &Error{Code: -70002, Message: "Number of tokens you want to buy larger than remaining tokens."}
	ErrNumberSmallerThanOwned     = &Error{Code: -70003, Message: "Number of tokens you want to sell larger than your owned tokens."}
	ErrIncorrectPrice             = &Error{Code: -70004, Message: "Price is incorrect."}
	ErrPaymentTransactionNotFound = &Error{Code: -70005, Message: "Payment transaction not found."}
	ErrTransactionNotFound        = &Error{Code: -70006, Message: "Transaction not found."}
	ErrGetBalance                 = &Error{Code: -70007, Message: "Can't get user's balance."}
	ErrNotEnoughMoney             = &Error{Code: -70008, Message: "Player doesn't have enough money."}
	ErrTransactionMismatch        = &Error{Code: -70009, Message: "Transaction information is mismatch."}
	ErrTransactionCompleted       = &Error{Code: -70010, Message: "Transaction is completed."}
	ErrPlayerPayForJail           = &Error{Code: -70011, Message: "Player must pay to get out of the jail."}
	ErrPlayerWrongPosition        = &Error{Code: -70011, Message: "Player can't buy the token at current position."}

	// stake:
	PaymenAddressAlreadyStaked = &Error{Code: -80000, Message: "The address is already staked"}
	PaymenAddressNotFound      = &Error{Code: -80001, Message: "The address is not found"}
	TxAlreadyStaked            = &Error{Code: -80002, Message: "The tx hash is already added"}
	AmountInvalid              = &Error{Code: -80003, Message: "The Amount invalid"}
	CanNotContributeNow        = &Error{Code: -80004, Message: "Can not contribute now!"}
	CanNotStopContributeNow    = &Error{Code: -80005, Message: "Can not stop contribute now!"}
	TxInvalid                  = &Error{Code: -80006, Message: "The tx hash invalid"}
	TxAlreadyStop              = &Error{Code: -80007, Message: "The address is already stop"}
	ProductNotFound            = &Error{Code: -80008, Message: "This QR code is unfamiliar. Please try again."}
	QRCodeAlreadyStaked        = &Error{Code: -80009, Message: "The qrcode is already staked"}
	AddrsesNotStake            = &Error{Code: -80010, Message: "The address haven't staked yet"}
	AlreadyRequested           = &Error{Code: -80011, Message: "The address is already requested"}

	ErrStakingOrderRewardClaimInvalidBalance   = &Error{Code: -80012, Message: "Reward balance is not enough"}
	ErrStakingOrderRewardClaimInvalidMinAMount = &Error{Code: -80013, Message: "Reward balance must be > Min amount"}
	DataInvalid                                = &Error{Code: -80014, Message: "The data invalid"}

	// order
	OrderAlreadyExists = &Error{Code: -90000, Message: "Order already exists"}
	AccessTokenDenied  = &Error{Code: -90001, Message: "Access token denied"}
	AuthorizeFailed    = &Error{Code: -90002, Message: "Authorize failed"}
	CaptureFailed      = &Error{Code: -90003, Message: "Capture failed"}
	UpdateOrderFailed  = &Error{Code: -90004, Message: "Update order failed"}
	AuthIDNotFound     = &Error{Code: -90005, Message: "AuthID Not Found"}

	// zelle:
	OrderIDNotFound          = &Error{Code: -90006, Message: "OrderIDNotFound"}
	OrderAlreadyUpdate       = &Error{Code: -90007, Message: "OrderAlreadyUpdate"}
	NotPermissonCode         = &Error{Code: -90008, Message: "NotPermissonCode"}
	UserDontSendEnoughtMoney = &Error{Code: -9009, Message: "UserDontSendEnoughtMoney"}
	CanNotParseParam         = &Error{Code: -90010, Message: "CanNotParseParam"}
	EmptyParamPrice          = &Error{Code: -90011, Message: "EmptyParamPrice"}
	EmptyParamOrderID        = &Error{Code: -90012, Message: "EmptyParamOrderID"}

	AmazonReject        = &Error{Code: -90013, Message: "AmazonReject"}
	AmazonTimeOutReject = &Error{Code: -90014, Message: "AmazonTimeOutReject"}
	AmazonPaymentFail   = &Error{Code: -90015, Message: "AmazonPaymentFail"}

	CardInvalid                    = &Error{Code: -90016, Message: "Your credit card information is invalid."}
	CreateCustomerFaied            = &Error{Code: -90017, Message: "Create customer faied"}
	CreateTransactionRequestFailed = &Error{Code: -90018, Message: "Create transaction request failed"}

	LimitOrder      = &Error{Code: -90020, Message: "Limit order now"}
	QuantityInvalid = &Error{Code: -90021, Message: "Quantity is invalid. 0< Quantity <= 10"}

	// Validator:
	ValidatorNotFound         = &Error{Code: -100000, Message: "Validator not found"}
	ErrValidatorAlreadyExists = &Error{Code: -100001, Message: "Validator id already exists"}

	ErrWalletAddressAlreadyExists = &Error{Code: -100002, Message: "Your address id already exists"}
	CanNotTransferNow             = &Error{Code: -100003, Message: "Can not trasnfer now!"}

	ErrBadgeAlreadyExists   = &Error{Code: -100004, Message: "Badge id already exists"}
	ErrNotificationNotFound = &Error{Code: -100005, Message: "Notification not found"}

	ErrTxTimeout                             = &Error{Code: -110001, Message: "Tx timeout"}
	ErrBalancetimeout                        = &Error{Code: -110002, Message: "Balance timeout"}
	ErrTxRejected                            = &Error{Code: -110003, Message: "Tx is rejected"}
	ErrTxNotSuccess                          = &Error{Code: -110003, Message: "Tx is not success"}
	ErrBalanceWrong                          = &Error{Code: -110003, Message: "Balance is wrong"}
	ErrGetTxAmount                           = &Error{Code: -110004, Message: "Can not get transactiona amount"}
	ErrDepositNotFound                       = &Error{Code: -110005, Message: "Deposit not found"}
	ErrAddressWrong                          = &Error{Code: -110006, Message: "Address wrong"}
	ErrTokenID                               = &Error{Code: -110007, Message: "TokenID is wrong"}
	ErrNetworkFeeTokenID                     = &Error{Code: -110008, Message: "NetworkFeeTokenID is wrong"}
	ErrBuyTokenID                            = &Error{Code: -110009, Message: "BuyTokenID is wrong"}
	ErrTokenIDAndNetworkFeeTokenID           = &Error{Code: -110010, Message: "TokenID and NetworkFeeTokenID is wrong"}
	ErrTokenIDAndAddress                     = &Error{Code: -110011, Message: "TokenID and Address is wrong"}
	ErrAddressAndNetworkFeeTokenID           = &Error{Code: -110012, Message: "Address and NetworkFeeTokenID is wrong"}
	ErrTokenIDAndNetworkFeeTokenIDAndAddress = &Error{Code: -110013, Message: "TokenID and NetworkFeeTokenID and Address is wrong"}

	// pAPP:
	TextMaxInvalid     = &Error{Code: -2000, Message: "This text is too long"}
	TextMinInvalid     = &Error{Code: -2001, Message: "This text is too long"}
	PAppNotFound       = &Error{Code: -2002, Message: "PApp not found"}
	CanNotAddPAppLogo  = &Error{Code: -2003, Message: "Can not add pApp logo"}
	CanNotAddPAppImage = &Error{Code: -2004, Message: "Can not add pApp image"}
	LogoRequired       = &Error{Code: -2005, Message: "Logo is required"}
	ImageRequired      = &Error{Code: -2006, Message: "Images required laset 1 image"}
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}
