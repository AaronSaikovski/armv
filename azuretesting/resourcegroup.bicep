
targetScope = 'subscription'

@description('Required. The name of the Resource Group.')
param name string = 'subscription-scoped-demo'

@description('Optional. Location of the Resource Group. It uses the deployment\'s location when not provided.')
param location string = deployment().location

@description('Optional. Tags of the resource group.')
param tags object = {}


resource resourceGroup 'Microsoft.Resources/resourceGroups@2021-04-01' = {
  location: location
  name: name
  tags: tags
  properties: {}
}
