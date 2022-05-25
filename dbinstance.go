package rdsnap

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

var Now = time.Now

type rdsClient struct {
	svc rdsiface.RDSAPI
}

func (r *rdsClient) createDBSnapshot(instanceId string) (string, error) {
	snapshotId := dbSnapshotId(instanceId)

	_, err := r.svc.CreateDBSnapshot(&rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: aws.String(instanceId),
		DBSnapshotIdentifier: aws.String(snapshotId),
	})
	if err != nil {
		return "", err
	}

	for {
		response, err := r.svc.DescribeDBSnapshots(&rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: aws.String(snapshotId),
		})
		if err != nil {
			return "", err
		}

		if *response.DBSnapshots[0].Status == "available" {
			break
		} else {
			time.Sleep(60 * time.Second)
		}
	}

	return snapshotId, nil
}

func (r *rdsClient) restoreDBInstanceFromDBSnapshot(cfg config, snapshotId string) (*config, error) {
	db, err := r.descDBInstance(cfg.instanceId)
	if err != nil {
		return nil, err
	}

	response, err := r.svc.RestoreDBInstanceFromDBSnapshot(&rds.RestoreDBInstanceFromDBSnapshotInput{
		CopyTagsToSnapshot:              db.CopyTagsToSnapshot,
		DBInstanceIdentifier:            aws.String(cfg.instanceId + "-snapshot"),
		DBParameterGroupName:            dbParamGroupName(db),
		DBSnapshotIdentifier:            aws.String(snapshotId),
		DBSubnetGroupName:               dbSubnetGroupName(db),
		DeletionProtection:              db.DeletionProtection,
		Domain:                          dbDomain(db),
		DomainIAMRoleName:               dbDomainIAMRoleName(db),
		EnableCloudwatchLogsExports:     db.EnabledCloudwatchLogsExports,
		EnableCustomerOwnedIp:           db.CustomerOwnedIpEnabled,
		EnableIAMDatabaseAuthentication: db.IAMDatabaseAuthenticationEnabled,
		OptionGroupName:                 dbOptionGroupName(db),
		PubliclyAccessible:              db.PubliclyAccessible,
		Tags:                            dbTagList(db),
		VpcSecurityGroupIds:             dbVpcSecurityGroupIds(db),
	})
	if err != nil {
		return nil, err
	}

	restoreDBInstanceId := *response.DBInstance.DBInstanceIdentifier

	rescfg := &config{}

	for {
		resdb, err := r.descDBInstance(restoreDBInstanceId)
		if err != nil {
			return nil, err
		}

		if *resdb.DBInstanceStatus == "available" {
			rescfg.instanceId = *resdb.DBInstanceIdentifier
			rescfg.engine = *resdb.Engine
			rescfg.host = *resdb.Endpoint.Address
			rescfg.port = *resdb.Endpoint.Port
			rescfg.user = cfg.user
			rescfg.password = cfg.password
			rescfg.dbtables = cfg.dbtables
			break
		} else {
			time.Sleep(60 * time.Second)
		}
	}

	return rescfg, nil
}

func (r *rdsClient) deleteDBInstance(instanceId string) error {
	response, err := r.svc.DeleteDBInstance(&rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
		SkipFinalSnapshot:    aws.Bool(true),
	})
	if err != nil {
		return err
	}

	if *response.DBInstance.DBInstanceStatus == "deleting" {
		return nil
	}

	return fmt.Errorf("Faild to delete DB instance.")
}

func (r *rdsClient) deleteDBSnapshot(snapshotId string) error {
	res, err := r.svc.DeleteDBSnapshot(&rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(snapshotId),
	})
	if err != nil {
		return err
	}

	if *res.DBSnapshot.Status == "deleted" {
		return nil
	}

	return fmt.Errorf("Faild to delete DB snapshot.")
}

func (r *rdsClient) descDBInstance(instanceId string) (*rds.DBInstance, error) {
	db, err := r.svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceId),
	})
	if err != nil {
		return nil, err
	}

	return db.DBInstances[0], nil
}

func dbSnapshotId(id string) string {
	return fmt.Sprintf("%s-%s", id, Now().Format("20060102-150405"))
}

func dbParamGroupName(dbInstance *rds.DBInstance) *string {
	return dbInstance.DBParameterGroups[0].DBParameterGroupName
}

func dbSubnetGroupName(dbInstance *rds.DBInstance) *string {
	return dbInstance.DBSubnetGroup.DBSubnetGroupName
}

func dbDomain(dbInstance *rds.DBInstance) *string {
	if len(dbInstance.DomainMemberships) == 0 {
		return nil
	}

	return dbInstance.DomainMemberships[0].Domain
}

func dbDomainIAMRoleName(dbInstance *rds.DBInstance) *string {
	if len(dbInstance.DomainMemberships) == 0 {
		return nil
	}

	return dbInstance.DomainMemberships[0].IAMRoleName
}

func dbOptionGroupName(dbInstance *rds.DBInstance) *string {
	return dbInstance.OptionGroupMemberships[0].OptionGroupName
}

func dbTagList(dbInstane *rds.DBInstance) []*rds.Tag {
	var tagList []*rds.Tag
	for _, tag := range dbInstane.TagList {
		if !strings.HasPrefix(*tag.Key, "aws:") {
			tagList = append(tagList, tag)
		}
	}

	return tagList
}

func dbVpcSecurityGroupIds(dbInstance *rds.DBInstance) []*string {
	var vpcSecurityGroupIds []*string
	for _, vsm := range dbInstance.VpcSecurityGroups {
		vpcSecurityGroupIds = append(vpcSecurityGroupIds, vsm.VpcSecurityGroupId)
	}

	return vpcSecurityGroupIds
}
