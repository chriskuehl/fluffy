// extensions which have icons available
var ICON_EXTENSIONS = [
	"7z", "ai", "bmp", "doc", "docx", "gif", "gz", "html",
	"jpeg", "jpg", "midi", "mp3", "odf", "odt", "pdf", "png", "psd", "rar",
	"rtf", "tar", "txt", "wav", "xls", "zip"
];

var fileIndex = 0;
var allFiles = [];
var uploading = false;

$(document).ready(function() {
	// browse button
	$("#selectFile").click(function() {
		input = addHiddenFileInput();
		input.trigger("click");
	});

	$("#selectFileInput").change(function() {
		console.log($(this).val());
	});

	// drop zone file drag-and-drop
	addDropZoneInput();

	// uploading
	$("#upload").click(function() {
		if (! uploading) {
			upload();
		}
	});
});

/**
 * Upload the queued files via XHR. Takes care of updating the UI (displaying
 * progress, hiding remove buttons, etc.)
 */
function upload() {
	if (! canUpload()) {
		return alert("Can't upload right now.");
	}

	uploading = true;

	// update UI
	var uploadButton = $("#upload");
	uploadButton.text("Cancel Upload");
	uploadButton.css("margin-bottom", "0px");

	$("#selectFiles").slideUp(200);
	$(".remove").fadeOut(200);

	// start the upload
	// http://stackoverflow.com/a/8244082
	$.ajax({
		url: "/upload",
		type: "POST",
		contentType: false,

		data: getFormData(),
		processData: false,

		xhr: function() {
			var req = $.ajaxSettings.xhr();

			req.upload.addEventListener("progress", function(e) {
				if (e.lengthComputable) {
					var percent = e.loaded / e.total;
					console.log(e.loaded + "\t" + e.total + "\t" + percent);
				}
			}, false);

			return req;
		},

		success: function(data) {
			console.log("returned: " + data);
		}
	});
}

/**
 * @return FormData object containing files to be uploaded
 */
function getFormData() {
	var formData = new FormData();
	formData.append("csrfmiddlewaretoken", $("#csrf_token").val());

	for (var i = 0; i < allFiles.length; i ++) {
		formData.append("file", allFiles[i]);
	}

	return formData;
}

/**
 * Takes a file input element and handles displaying the file to the user and
 * storing the file to be submitted with the request.
 *
 * The input will be removed from the DOM.
 *
 * @param input - jQuery input object
 */
function handleInput(input) {
	if (uploading) {
		return;
	}

	var files = input[0].files;

	for (var i = 0; i < files.length; i ++) {
		var file = files[i];

		if (! fileAlreadyQueued(file)) {
			allFiles.push(file);
			displayFile(file);
		}
	}

	input.remove();
}

/**
 * Checks if a file is already queued to be uploaded, checking for duplicates
 * based on name.
 *
 * @return whether or not a file is already queued to be uploaded
 */
function fileAlreadyQueued(file) {
	for (var i = 0; i < allFiles.length; i ++) {
		if (allFiles[i].name == file.name) {
			return true;
		}
	}

	return false;
}

/**
 * Displays a file in the file list.
 *
 * @param file - File object
 */
function displayFile(file) {
	var li = $("<li />");

	var progress = $("<div />");
	progress.addClass("progress");
	progress.appendTo(li);

	var icon = $("<img />");
	icon.attr("src", getIcon(file.name));
	icon.appendTo(li);

	var title = $("<div />");
	title.addClass("title");
	title.text(file.name);
	title.appendTo(li);

	var remove = $("<a />");
	remove.addClass("remove");
	remove.html("&times;");
	remove.appendTo(li);

	remove.click(function() {
		if (uploading) {
			return;
		}

		var idx = allFiles.indexOf(file);

		if (idx > -1) {
			allFiles.splice(idx, 1);
		}

		li.remove();
		updateUpload();
	});

	li.appendTo($("#files"));
	updateUpload();
}

/**
 * Show or hide the upload button based on whether or not any files are
 * queued for upload.
 */
function updateUpload() {
	var visible = $("#upload").is(":visible");
	var shouldBeVisible = canUpload();

	if (visible && ! shouldBeVisible) {
		$("#upload").hide();
	} else if (! visible && shouldBeVisible) {
		// set display -> block instead of calling show() since otherwise
		// jQuery will "restore" it to inline
		$("#upload").css("display", "block");
	}
}

/**
 * @return whether or not an upload can proceed
 */
function canUpload() {
	return allFiles.length > 0;
}

/**
 * Returns the path of the icon for a file.
 *
 * @param fileName
 * @return relative path to the icon
 */
function getIcon(fileName) {
	var parts = fileName.split(".");
	var extension = parts[parts.length - 1].toLowerCase();

	if (ICON_EXTENSIONS.indexOf(extension) == -1) {
		extension = "unknown";
	}

	return "/static/img/mime/small/" + extension + ".png";
}

/**
 * Creates a new hidden file input in <body>. You probably want to call click()
 * on the return value to display the browse window.
 *
 * @return jQuery input object
 */
function addHiddenFileInput() {
	var input = makeFileInput();

	input.appendTo($("body"));
	input.addClass("hiddenFileInput");

	return input;
}

/**
 * Makes a new file input with a unique name which will call handleInput (and
 * thus take care of everything needed for files to be uploaded).
 */
function makeFileInput() {
	var input = $("<input type=\"file\" />");

	input.attr({
		name: "file-" + (fileIndex ++) + "[]",
		multiple: "multiple"
	});

	input.change(function() {
		handleInput(input);
	});

	return input;
}

/**
 * Create an input in #dropZone which can handle drag/drop file uploads
 */
function addDropZoneInput() {
	input = makeFileInput();

	input.appendTo($("#dropZone"));
	input.on("dragenter", dropZoneHoverOn);
	input.on("dragleave drop", dropZoneHoverOff);

	input.change(function() {
		addDropZoneInput();
	});
}

function dropZoneHoverOn() {
	$("#dropZone").css("backgroundColor", "#FFEECA");
}

function dropZoneHoverOff() {
	$("#dropZone").css("backgroundColor", "#F3FFE2");
}
