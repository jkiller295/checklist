package i18n

import "strings"

type Translations map[string]string

var locales = map[string]Translations{
	"en": {
		"app_name":            "Checklist",
		"sign_in":             "Sign In",
		"password":            "Password",
		"wrong_password":      "Wrong password. Try again.",
		"sign_out":            "Sign out",
		"new_list":            "New list",
		"list_name":           "List name",
		"create":              "Create",
		"cancel":              "Cancel",
		"delete":              "Delete",
		"rename":              "Rename",
		"add_item":            "Add item",
		"item_placeholder":    "What needs doing?",
		"empty_list":          "Nothing here yet. Add something above.",
		"all_done":            "All done! 🎉",
		"lists":               "Lists",
		"back":                "Back",
		"save":                "Save",
		"delete_list_confirm": "Delete this list and all its items?",
		"delete_item_confirm": "Delete this item?",
		"items_done":          "done",
		"language":            "Language",
		"select_all":          "Select all",
		"no_lists":            "No lists yet. Create one to get started.",
		"edit":                "Edit",
		"delete_checked":      "Delete checked items",
		"uncheck_all":         "Uncheck all",
		"backup":              "Backup",
		"restore":             "Restore",
		"restore_confirm":     "This will overwrite all current lists and items. Continue?",
		"backup_restore":      "Backup & Restore",
		"backup_restore_desc": "Download a backup of your checklist data or restore from a previous backup file.",
		"backup_desc":         "Export all lists and items to a JSON file.",
		"download_backup":     "Download Backup",
		"restore_desc":        "Import a previously exported backup file.",
		"restore_warning":     "Restoring will overwrite all current lists and items.",
		"restore_now":         "Restore Now",
	},
	"vi": {
		"app_name":            "Danh sách",
		"sign_in":             "Đăng nhập",
		"password":            "Mật khẩu",
		"wrong_password":      "Sai mật khẩu. Thử lại.",
		"sign_out":            "Đăng xuất",
		"new_list":            "Danh sách mới",
		"list_name":           "Tên danh sách",
		"create":              "Tạo",
		"cancel":              "Hủy",
		"delete":              "Xóa",
		"rename":              "Đổi tên",
		"add_item":            "Thêm mục",
		"item_placeholder":    "Cần làm/mua gì?",
		"empty_list":          "Chưa có gì. Thêm mục ở trên.",
		"all_done":            "Xong hết rồi! 🎉",
		"lists":               "Danh sách",
		"back":                "Quay lại",
		"save":                "Lưu",
		"delete_list_confirm": "Xóa danh sách và tất cả mục?",
		"delete_item_confirm": "Xóa mục này?",
		"items_done":          "xong",
		"language":            "Ngôn ngữ",
		"select_all":          "Chọn tất cả",
		"no_lists":            "Chưa có danh sách. Tạo một danh sách để bắt đầu.",
		"edit":                "Sửa",
		"delete_checked":      "Xóa các mục đã chọn",
		"uncheck_all":         "Bỏ chọn tất cả",
		"backup":              "Sao lưu",
		"restore":             "Khôi phục",
		"restore_confirm":     "Thao tác này sẽ ghi đè toàn bộ danh sách và mục hiện tại. Tiếp tục?",
		"backup_restore":      "Sao lưu & Khôi phục",
		"backup_restore_desc": "Tải xuống bản sao lưu dữ liệu danh sách của bạn hoặc khôi phục từ tệp sao lưu trước đó.",
		"backup_desc":         "Xuất tất cả danh sách và mục vào một tệp JSON.",
		"download_backup":     "Tải file Sao lưu xuống",
		"restore_desc":        "Nhập một tệp sao lưu đã xuất trước đó.",
		"restore_warning":     "Việc khôi phục sẽ ghi đè toàn bộ danh sách và mục hiện tại.",
		"restore_now":         "Khôi phục ngay",
	},
}

func T(lang, key string) string {
	lang = normalize(lang)
	if t, ok := locales[lang]; ok {
		if v, ok := t[key]; ok {
			return v
		}
	}
	// fallback to English
	if t, ok := locales["en"]; ok {
		if v, ok := t[key]; ok {
			return v
		}
	}
	return key
}

func normalize(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	// Accept "vi", "vi-VN", etc.
	if strings.HasPrefix(lang, "vi") {
		return "vi"
	}
	return "en"
}

func ValidLang(lang string) bool {
	_, ok := locales[normalize(lang)]
	return ok
}

func Normalize(lang string) string {
	return normalize(lang)
}
