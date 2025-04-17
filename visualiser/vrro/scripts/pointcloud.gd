extends MeshInstance3D

var cloud_mat = preload("res://materials/point_material.tres")
@onready var http = $"../CloudRequest"
var ready_to_request = true
@export var eraseCloud : String = "erase"
var st : SurfaceTool
var cloud : PackedVector3Array


# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	st = SurfaceTool.new()
	st.begin(Mesh.PRIMITIVE_POINTS)
	http.request_completed.connect(self._cloud_request_completed)
	

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta: float) -> void:
	
	if Input.is_action_just_pressed(eraseCloud):
		self.clear()
		
	if ready_to_request:
		download_cloudpoint()
		
func clear() -> void:
	st.clear()
	self.mesh = st.commit()
	st.begin(Mesh.PRIMITIVE_POINTS)
		

func points_to_mesh(jsonPoints: String) -> void:
	if jsonPoints == null:
		return
		
	var newCloud = JSON.parse_string(jsonPoints)
	
	if newCloud == null:
		return
	
	st.clear()
	st.begin(Mesh.PRIMITIVE_POINTS)
	for newPoint in newCloud:
		st.add_vertex(Vector3(newPoint.x, newPoint.y, newPoint.z))

	self.mesh = st.commit()

func exorcise_phantoms(point) -> void:
	print(self.mesh.get_faces())


func download_cloudpoint() -> void:
	ready_to_request = false
	
	# Perform a GET request. The URL below returns JSON as of writing.
	var error = http.request("http://localhost:8081/cloudpoints")
	if error != OK:
		push_error("An error occurred in the HTTP request.")

# Called when the HTTP request is completed.
func _cloud_request_completed(result, response_code, headers, body):
	# Will print the user agent string used by the HTTPRequest node (as recognized by httpbin.org).
	points_to_mesh(body.get_string_from_utf8())
	ready_to_request = true
	
