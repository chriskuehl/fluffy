$(document).ready(function() {
	$("#selectFile").click(function() {
		$("#selectFileInput").trigger("click");
	});

	$("#selectFileInput").change(function() {
		console.log($(this).val());
	});
});
