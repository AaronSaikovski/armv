param storageLocation string

@description('The name of the Storage Account')
param storageName string = 'stg${uniqueString(resourceGroup().id)}'


resource storageAcct 'Microsoft.Storage/storageAccounts@2023-01-01' = {
  name: storageName
  location: storageLocation
  sku: {
    name: 'Standard_LRS'
  }
  kind: 'Storage'
  properties: {}
}
