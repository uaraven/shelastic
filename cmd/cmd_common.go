package cmd

import (
	flags "github.com/jessevdk/go-flags"
)

var (
	errIndexNotSelected = "No index selected. Select index using 'use <index-name>' command or by passing --index <index-name> parameter"
)

type documentSelectorData struct {
	Index    string `long:"index" description:"Index name"`
	Document string `long:"doc" description:"Document type"`
	Args     []string
}

type documentSelector interface {
	GetIndex() string
	GetDocument() string
	GetArgs() []string

	SetIndex(p string)
	SetDocument(p string)
	SetArgs(p []string)
}

func (ds *documentSelectorData) GetIndex() string {
	return ds.Index
}

func (ds *documentSelectorData) GetDocument() string {
	return ds.Document
}

func (ds *documentSelectorData) GetArgs() []string {
	return ds.Args
}

func (ds *documentSelectorData) SetIndex(index string) {
	ds.Index = index
}

func (ds *documentSelectorData) SetDocument(doc string) {
	ds.Document = doc
}

func (ds *documentSelectorData) SetArgs(args []string) {
	ds.Args = args
}

func parseDocumentArgs(args []string) (*documentSelectorData, error) {
	res, err := parseDocumentArgsCustom(args, &documentSelectorData{})
	if err != nil {
		return nil, err
	}
	return res.(*documentSelectorData), nil
}

func parseDocumentArgsCustom(args []string, customOpts interface{}) (interface{}, error) {
	positional, err := flags.ParseArgs(customOpts, args)
	if err != nil {
		return nil, err
	}

	opts := customOpts.(documentSelector)
	if opts.GetIndex() == "" {
		opts.SetIndex(context.ActiveIndex)
	}
	opts.SetArgs(positional)

	return customOpts, nil
}
