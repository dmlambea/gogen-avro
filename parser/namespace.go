package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/actgardner/gogen-avro/schema"
)

const (
	// Special Go type for null types
	nullGoType = ""
)

// Namespace holds an array of schema.ComplexType, used to create the .go source files from them.
type Namespace struct {
	registry referenceRegistry
	Roots    []schema.ComplexType
	// shortUnions    bool
}

func NewNamespace(shortUnions bool) *Namespace {
	return &Namespace{
		registry: NewReferenceRegistry(),
		// shortUnions: shortUnions,
	}
}

// ParseSchema accepts an Avro schema as a JSON string and parses it. An error is returned if any occurs.
// The Avro type defined at the top level and all the type definitions beneath it will also be added to this Namespace.
func (n *Namespace) ParseSchema(schemaJson []byte) (topLevel schema.GenericType, err error) {
	var schema interface{}
	if err = json.Unmarshal(schemaJson, &schema); err != nil {
		return
	}

	// Intentionally the name starts with an invalid character, so the identifier cannot collide
	// with any user-defined one.
	return n.decodeType("-", "", schema)
}

// registerType creates a reference wrapping a qnamed type in the registry. To avoid post-processing,
// all qnamed types that are complex types (i.e., all non-primitive types that might be requiring a Go
// source file), are also registered as root types. The registry triggers the resolution of all
// references that were used in other types before reading the real definition.
func (n *Namespace) registerType(qname schema.QName, t schema.GenericType) schema.GenericType {
	if ct, ok := t.(schema.ComplexType); ok {
		n.Roots = append(n.Roots, ct)
	}
	ref, err := n.registry.CreateReference(qname, t)
	if err != nil {
		panic(err)
	}
	if ref.IsUntyped() {
		return ref
	}
	return ref.Type()
}

// decodeType is the generic method to decode any type, given its schema.
func (n *Namespace) decodeType(name, namespace string, typeSchema interface{}) (schema.GenericType, error) {
	switch t := typeSchema.(type) {
	case map[string]interface{}:
		return n.decodeComplexType(name, namespace, typeSchema.(map[string]interface{}))
	case []interface{}:
		// Unions must be manually added to the roots, because they are complex types
		// but are not namespaced, so unions never go to the registry.
		union, err := n.decodeUnion(name, namespace, typeSchema.([]interface{}))
		if err == nil {
			// In case of getting a valid union
			if ct, ok := union.(schema.ComplexType); ok {
				n.Roots = append(n.Roots, ct)
			}
		}
		return union, err
	case string:
		name = typeSchema.(string)
		return n.getTypeByName(name, namespace), nil
	default:
		return nil, fmt.Errorf("decoding of type %s is unimplemented", t)
	}
}

// decodeComplexType decodes and registers a fixed, enum, map, array, record or qnamed reference complex type
func (n *Namespace) decodeComplexType(name, namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	typeStr, err := stringFromMap(schemaMap, "type")
	if err != nil {
		return nil, err
	}
	var at schema.GenericType
	switch typeStr {
	case "enum":
		at, err = n.decodeEnum(namespace, schemaMap)
	case "fixed":
		at, err = n.decodeFixed(namespace, schemaMap)
	case "map":
		at, err = n.decodeMap(namespace, schemaMap)
	case "array":
		at, err = n.decodeArray(namespace, schemaMap)
	case "record":
		at, err = n.decodeRecord(namespace, schemaMap)
	default:
		// If the type isn't a special case, it's a primitive or a reference to an existing type
		return n.getTypeByName(typeStr, namespace), nil
	}
	if err != nil {
		return nil, err
	}
	if ns, ok := at.(schema.NamespacedType); ok {
		at = n.registerType(ns.QName(), at)
	}
	return at, nil
}

// decodeEnum accepts a namespace and a map representing an enum definition,
// it validates the definition and build the enum type struct.
func (n *Namespace) decodeEnum(namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	qname, err := extractQName(schemaMap, namespace)
	if err != nil {
		return nil, err
	}

	symbolsSlice, err := arrayFromMap(schemaMap, "symbols")
	if err != nil {
		return nil, err
	}

	symbols, ok := interfaceSliceToStringSlice(symbolsSlice)
	if !ok {
		return nil, errors.New("'symbols' must be an array of strings")
	}

	enumField := schema.NewEnumField(qname, symbols)

	if err = parseAliases(schemaMap, enumField, qname.Namespace); err != nil {
		return nil, err
	}
	if err = parseDoc(schemaMap, enumField); err != nil {
		return nil, err
	}

	return enumField, nil
}

// decodeFixed accepts a namespace and a map representing a fixed type definition,
// it validates the definition and build the fixed type struct.
func (n *Namespace) decodeFixed(namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	qname, err := extractQName(schemaMap, namespace)
	if err != nil {
		return nil, err
	}

	sizeBytes, err := floatFromMap(schemaMap, "size")
	if err != nil {
		return nil, err
	}
	if sizeBytes < 0 {
		return nil, errors.New("'size' must be a positive integer")
	}

	fixedField := schema.NewFixedField(qname, uint64(sizeBytes))

	if err = parseAliases(schemaMap, fixedField, qname.Namespace); err != nil {
		return nil, err
	}

	return fixedField, nil
}

// decodeMap accepts a namespace and a map representing a fixed type definition,
// it validates the definition and build the map type struct.
func (n *Namespace) decodeMap(namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	values, ok := schemaMap["values"]
	if !ok {
		return nil, NewRequiredMapKeyError("values")
	}

	itemType, err := n.decodeType("", namespace, values)
	if err != nil {
		return nil, err
	}

	return schema.NewMapField(itemType), nil
}

// decodeArray accepts a namespace and a map representing a fixed type definition,
// it validates the definition and build the array type struct.
func (n *Namespace) decodeArray(namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	items, ok := schemaMap["items"]
	if !ok {
		return nil, NewRequiredMapKeyError("items")
	}

	itemType, err := n.decodeType("", namespace, items)
	if err != nil {
		return nil, err
	}

	return schema.NewArrayField(itemType), nil
}

// decodeRecord accepts a namespace and a map representing a fixed type definition,
// it validates the definition and build the record type struct.
func (n *Namespace) decodeRecord(namespace string, schemaMap map[string]interface{}) (schema.GenericType, error) {
	qname, err := extractQName(schemaMap, namespace)
	if err != nil {
		return nil, err
	}

	fields, err := arrayFromMap(schemaMap, "fields")
	if err != nil {
		return nil, err
	}

	decodedFields := make([]schema.GenericType, len(fields))
	for i, f := range fields {
		fieldSchemaMap, ok := f.(map[string]interface{})
		if !ok {
			return nil, NewWrongMapValueTypeError("fields", "map[]", fields)
		}

		fieldName, err := stringFromMap(fieldSchemaMap, "name")
		if err != nil {
			return nil, err
		}

		fieldType, err := n.decodeType(fieldName, qname.Namespace, fieldSchemaMap["type"])
		if err != nil {
			return nil, err
		}

		decodedField := schema.NewField(fieldName, fieldType, i)

		// Record fields have no namespaces
		if err = parseAliases(fieldSchemaMap, decodedField, ""); err != nil {
			return nil, err
		}
		if err = parseDoc(fieldSchemaMap, decodedField); err != nil {
			return nil, err
		}

		decodedFields[i] = decodedField

		/**  TODO: support golang tags
		var fieldTags string
		if tags, ok := field["golang.tags"]; ok {
			fieldTags, ok = tags.(string)
			if !ok {
				return nil, NewWrongMapValueTypeError("golang.tags", "string", tags)
			}
		}
		*/
	}

	recordField := schema.NewRecordField(qname, decodedFields)

	if err = parseAliases(schemaMap, recordField, qname.Namespace); err != nil {
		return nil, err
	}
	if err = parseDoc(schemaMap, recordField); err != nil {
		return nil, err
	}

	recordField.SetSchema(schemaMap)

	return recordField, nil
}

// decodeUnion accepts a namespace and a map representing a union type definition,
// it validates the definition and build the union type struct.
func (n *Namespace) decodeUnion(name, namespace string, schemaList []interface{}) (result schema.GenericType, err error) {
	nullFieldFoundAt := -1
	decodedFields := make([]schema.GenericType, len(schemaList))
	for i, f := range schemaList {
		fieldType, err := n.decodeType("", "", f)
		if err != nil {
			return nil, err
		}
		if fieldType.GoType() == nullGoType {
			nullFieldFoundAt = i
		}

		decodedFields[i] = schema.NewField(fieldType.Name(), fieldType, i)
	}

	var u *schema.UnionType
	if u, err = schema.NewUnionField(decodedFields), nil; err == nil {
		u.SetOptionalIndex(nullFieldFoundAt)
		result = u
	}
	return
}

// getTypeByName returns the type associated with a type name, mostly primitive types, but also qnamed, registered types.
func (n *Namespace) getTypeByName(typeStr, namespace string) schema.GenericType {
	if t := schema.NewPrimitiveType(typeStr); t != nil {
		return t
	}
	// Non-primitive type: create a reference
	qname := fixQName(schema.QName{Name: typeStr, Namespace: namespace})
	return n.registerType(qname, nil)
}

// Enriches a NamespacedType with aliases data, if any
func parseAliases(schemaMap map[string]interface{}, t schema.NamespacedType, namespace string) error {
	aliases, ok := schemaMap["aliases"]
	if !ok {
		return nil
	}

	interfaceList, ok := aliases.([]interface{})
	if !ok {
		return fmt.Errorf("Field aliases expected to be array, got %v", aliases)
	}
	stringList, ok := interfaceSliceToStringSlice(interfaceList)
	if !ok {
		return fmt.Errorf("Field aliases expected to be array of string, got %v", interfaceList)
	}

	qnames := make([]schema.QName, 0, len(interfaceList))
	for _, alias := range stringList {
		qnameAlias := schema.QName{
			Name:      alias,
			Namespace: namespace,
		}
		qnames = append(qnames, fixQName(qnameAlias))
	}
	t.SetAliases(qnames)
	return nil
}

// Enriches a DocumentedType with doc string data, if declared
func parseDoc(schemaMap map[string]interface{}, t schema.DocumentedType) error {
	var docString string
	if doc, ok := schemaMap["doc"]; ok {
		if docString, ok = doc.(string); !ok {
			return errors.New("'doc' must be a string")
		}
		t.SetDoc(docString)
	}
	return nil
}

// extractQName returns the fully qualified name from the
// schema definition represented by schemaMap. From the specification:
//
//	In record, enum and fixed definitions, the fullname is
//	determined in one of the following ways:
//
//	- A name and namespace are both specified. For example, one
//	might use "name": "X", "namespace": "org.foo" to indicate the
//	fullname org.foo.X.
//
//	- A fullname is specified. If the name specified contains a
//	dot, then it is assumed to be a fullname, and any namespace
//	also specified is ignored. For example, use "name":
//	"org.foo.X" to indicate the fullname org.foo.X.
//
//	- A name only is specified, i.e., a name that contains no
//	dots. In this case the namespace is taken from the most
//	tightly enclosing schema or protocol. For example, if "name":
//	"X" is specified, and this occurs within a field of the record
//	definition of org.foo.Y, then the fullname is org.foo.X. If
//	there is no enclosing namespace then the null namespace is
//	used.
func extractQName(schemaMap map[string]interface{}, enclosing string) (qname schema.QName, err error) {
	if qname.Name, err = stringFromMap(schemaMap, "name"); err != nil {
		return
	}
	if _, ok := schemaMap["namespace"]; ok {
		if qname.Namespace, err = stringFromMap(schemaMap, "namespace"); err != nil {
			return
		}
	}
	if qname.Namespace == "" {
		qname.Namespace = enclosing
	}
	return fixQName(qname), nil
}

// fixQName parses a name according to the Avro spec:
//   - If the name contains a dot ('.'), the last part is the name and the rest is the namespace
//   - Otherwise, the enclosing namespace is used
func fixQName(qname schema.QName) schema.QName {
	if lastIndex := strings.LastIndex(qname.Name, "."); lastIndex != -1 {
		qname.Namespace = qname.Name[:lastIndex]
		qname.Name = qname.Name[lastIndex+1:]
	}
	return qname
}

func interfaceSliceToStringSlice(iSlice []interface{}) ([]string, bool) {
	var ok bool
	stringSlice := make([]string, len(iSlice))
	for i, v := range iSlice {
		stringSlice[i], ok = v.(string)
		if !ok {
			return nil, false
		}
	}
	return stringSlice, true
}
