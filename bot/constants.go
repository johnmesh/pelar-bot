package bot

const LOGIN_ACCOUNT_LINK_ID = "id_esauth_myaccount_login_link"
const LOGIN_EMAIL_ID = "id_esauth_login_field"
const LOGIN_PASSWORD_ID = "id_esauth_pwd_field"
const LOGIN_BUTTON = "id_esauth_login_button"

const AVAILABLE_ORDERS_TAB = "available_tab"
const ORDER_LIST = "available_orders_list_container"
const ORDER_CONTAINER = "order_container"
const ORDER_TYPE = "service_type"
const ORDER_NUMBER_VIEW = "order_number"
const ORDER_PAPER_INSTRUCTIONS = "paper_instructions_view"

const CUSTOMER_ONLINE_TITLE = "Customer online" //<a/>
const CUSTOMER_OFFLINE = "Customer offline"
const CUST_RATING = "customer-rating" //also contains completed orders 8.77 /74
const ORDER_DISCIPLINE = "discipline"
const NEW_CUSTOMER = "new-customer"
const ORDER_PAGES = "pagesamount"

const ORDER_DEADLINE = "td_deadline"
const ORDER_DEADLINE_DATE = "d-left"
const BIDDING_FORM = "id_order_bidding_form"
const BID_BTN = "apply_order"
const ORDER_LOADING = "empty_orders_list_message_container"
const TIMER = "id_read_timeout_sec"

const ORDER_DISCARD = "order_container_discard"
const BID_INPUT = "id_bid"
const BID_BUTTON_ID = "apply_order"
const BID_REC = "rec_bid"

var Amount = map[string]string{
	"4":  "4",
	"5":  "4",
	"6":  "4.8",
	"7":  "5",
	"8":  "6",
	"9":  "7",
	"10": "8",
	"11": "9",
	"12": "9",
	"13": "10",
	"14": "11",
}

var ExDiscipines = []string{
	"Engineering",
	"Account",
	"Industrial Engineering",
	"dance", "Accounting",
	"finance",
	"Electrical",
	"Financial Accounting",
	"Engineering : Architecture",
	"Chemistry", "Excel Mathematics",
	"mechanical Engineering",
	"Business and Management : Finance",
	"Tax", "Physics",
	"Engineering:engineering",
	"Natural science:chemistry",
	"Engineering:Technology",
	"Engineering : Architecture",
	"Mathematics and Statistics",
	"Mathematics and Statistics:Analysis",
	"Other:mechanical", "bio staticsmy",
	"chemical design project",
	"chemical prj", "financial",
	"Auditin",
	"Civil engineering",
	"environmental design module",
	"Engineering practice",
	"Management Sciences and Engineering",
	"Engineering & physics",
	"Masters of Engineering in Electrical Machines & Power Engineering",
	"Mathematics and Statistics : Mathematics",
	"Mathematics and Statistics : Statistics",
	"Engineering : Chemical Engineering",
	"Lab report", "Managerial Accounting",
	"Spss Lab", "Computer Science : MATLAB",
	"Discipline: Computer and Web Programming : Assembly Code Program",
	"Discipline: Database design and optimisation : SQL Programming",
	"Discipline: Computer and Web Programming : Computer Programming",
	"Discipline: Digital Design, UX and Visual Design : Graphic Design",
	"Computer and Web Programming : Computer Programming", "Computer Science : Computer Science",
	"Digital Design, UX and Visual Design : Graphic Design",
	"Computer and Web Programming : C Programming",
	"Computer and Web Programming : Python Programming",
	"Computer and Web Programming : C# Programming",
	"Computer and Web Programming : Web Programming",
	"Computer and Web Programming : Java Programming",
	"Database design and optimisation : MySQL Programming",
	"Computer and Web Programming : C++ Programming",
	"Computer and Web Programming : JavaScript",
}

//{"_id":{"$oid":"5f480869eb8840c19f507403"},"email":"aceicewriting@gmail.com","password":"3381","acc":{"type":"es","details":{"email":"lydiarugut@gmail.com","password":"my  shark"}},"bot-settings":{"bids":[{"rec":4,"bid":4},{"rec":5,"bid":4},{"rec":6,"bid":4.800000190734863},{"rec":7,"bid":5},{"rec":8,"bid":6},{"rec":9,"bid":7},{"rec":10,"bid":8},{"rec":11,"bid":9},{"rec":12,"bid":9},{"rec":13,"bid":10},{"rec":14,"bid":11}],"general":{"refresh_interval":2,"threads":3,"stop_time":{"$numberLong":"4000"},"urgency_period":2,"run_bg":true},"order":{"min_deadline":{"$numberLong":"21600"},"max_deadline":{"$numberLong":"31556952"},"min_pages":1,"max_pages":100,"max_urgency_pages":100,"complete_orders":0,"min_rating":9,"discard_offline_cust":false,"discard_assignments":true,"discard_editting":true,"discard_noratings":false,"discard_new_cust":false,"ex_discipline":["\u0001\u0001\u0001]}},"expires_in":"","product_key":"OHR5ZQ3RDL","bidAccs":[{"email":"jacknyangare@yahoo.com","password":"","productKey":"PD64J9IHKG","expiresIn":"20210506213653"},{"email":"onderidismus85@gmail.com","password":"","productKey":"PD64J9IHKG","expiresIn":"20210506180825"},{"email":"jacknyangare@yahoo.com","password":"","productKey":"OHV14TGBG5","expiresIn":"20201130181243"},{"email":"lydiarugut@gmail.com","password":"","productKey":"PC77FSELD1","expiresIn":"20210406181305"},{"email":"onderdismus85@gmail.com","password":"","productKey":"","expiresIn":""},{"email":"nambengeleashap@gmail.com","password":"","productKey":"PCG1QPVWQV","expiresIn":"20210415112332"},{"email":"onderidismus85gmail.com","password":"","productKey":"PD64J9IHKG","expiresIn":"20210508160431"}],"status":"active"}
