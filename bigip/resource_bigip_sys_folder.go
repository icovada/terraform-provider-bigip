/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"context"
	"fmt"
	"log"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBigipSysFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysFolderCreate,
		UpdateContext: resourceBigipSysFolderUpdate,
		ReadContext:   resourceBigipSysFolderRead,
		DeleteContext: resourceBigipSysFolderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Name of folder",
				ForceNew:    true,
			},
			"appService": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The application service that the object belongs to.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User-defined description of the folder",
			},
			"deviceGroup": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Associate this folder with a device failover group or device sync group. ‘default’ to associate this folder with its parent’s device group. ‘non-default’ to leave this field’s value untouched but disassociate this folder from its parent.",
			},
			"hidden": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "false",
				Description: "Specifies if this folder will be hidden. If set to ‘true’, this folder will be hidden from standard command usage. Administrators can display, modify or remove hidden folders using the ‘-hidden’ option.",
			},
			"noRefCheck": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Specifies whether strict device group reference validation is performed during sync behavior on items in this folder",
			},
			"trafficGroup": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Associate this folder with a network failover group. ‘default’ to associate this folder with its parent’s device group. ‘non-default’ to leave this field’s value untouched but disassociate this folder from its parent.",
			},
		},
	}

}

func resourceBigipSysFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)


	name := d.Get("name").(string)
	appService := d.Get("appService").(string)
	description := d.Get("description").(string)
	deviceGroup := d.Get("deviceGroup").(string)
	hidden := d.Get("hidden").(string)
	noRefCheck := d.Get("noRefCheck").(string)
	trafficGroup := d.Get("trafficGroup").(string)

	log.Printf("[INFO] Configuring Folder %s", name)
	folderConfig := &bigip.Folder{
		Name: name,
		AppService: appService,
		Description: description,
		DeviceGroup: deviceGroup,
		Hidden: hidden,
		NoRefCheck: noRefCheck,
		TrafficGroup: trafficGroup,
	}

	log.Printf("[DEBUG] config of Folder to be add :%+v", folderConfig)
	d.SetId(name)

	exists, _ := resourceSysFolderExists(d, meta)
	if !exists {
		if err := client.AddFolder(folderConfig); err != nil {
			d.SetId("")
			return diag.FromErr(fmt.Errorf("error creating Folder %s; %v", name, err))
		}
	}

	return resourceBigipSysFolderRead(ctx, d, meta)
}


func resourceSysFolderExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Fetching folder " + name)

	node, err := client.GetFolder(name)
	if err != nil {
		log.Printf("[ERROR] Unable to retrieve folder %s  %v :", name, err)
		return false, err
	}

	if node == nil {
		log.Printf("[WARN] folder (%s) not found, removing from state", d.Id())
		return false, nil
	}
	return true, nil
}



func resourceBigipSysFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Updating Folder " + name)
	folderConfig := &bigip.Folder{
		Name: d.Get("name").(string),
		AppService: d.Get("appService").(string),
		Description: d.Get("description").(string),
		DeviceGroup: d.Get("deviceGroup").(string),
		Hidden: d.Get("hidden").(string),
		NoRefCheck: d.Get("noRefCheck").(string),
		TrafficGroup: d.Get("trafficGroup").(string),
	}

	if err := client.ModifyFolder(name, folderConfig); err != nil{
		return diag.FromErr(fmt.Errorf("Error modifying Folder %s: %v", name, err))
	}
	return resourceBigipSysFolderRead(ctx, d, meta)
}

func resourceBigipSysFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Reading Folder  " + name)

	folder, err := client.GetFolder(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Folder (%v) ", err)
		return diag.FromErr(err)
	}
	if folder == nil {
		log.Printf("[WARN] Folder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	_ = d.Set("name", folder.Name)
	_ = d.Set("appService", folder.AppService)
	_ = d.Set("description", folder.Description)
	_ = d.Set("deviceGroup", folder.DeviceGroup)
	_ = d.Set("hidden", folder.Hidden)
	_ = d.Set("noRefCheck", folder.NoRefCheck)
	_ = d.Set("trafficGroup", folder.TrafficGroup)

	return nil
}

func resourceBigipSysFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	
	name := d.Id()
	log.Println("[INFO] Deleting Folder:" + name)

	err := client.DeleteFolder(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Folder (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}