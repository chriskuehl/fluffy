// extensions which have icons available
var ICON_EXTENSIONS = [
	"7z", "ai", "bmp", "doc", "docx", "gif", "gz", "html",
	"jpeg", "jpg", "midi", "mp3", "odf", "odt", "pdf", "png", "psd", "rar",
	"rtf", "tar", "txt", "wav", "xls", "zip"
];

var fileIndex = 0;
var allFiles = [];

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
});

/**
 * Takes a file input element and handles displaying the file to the user and
 * storing the file to be submitted with the request.
 *
 * The input will be removed from the DOM.
 *
 * @param input - jQuery input object
 */
function handleInput(input) {
	var files = input[0].files;

	for (var i = 0; i < files.length; i ++) {
		var file = files[i];

		allFiles.push(file);
		displayFile(file);
	}

	input.remove();
}

/**
 * Displays a file in the file list.
 *
 * @param file - File object
 */
function displayFile(file) {
	var li = $("<li />");

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

	li.appendTo($("#files"));
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
