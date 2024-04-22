targetScope = 'subscription'

param sourcersg string = 'rsg-source'
param destrsg string = 'rsg-destination'

param resourceGroupLocation string = 'australiaeast'

module sourceRsg 'resourcegroup.bicep' = {
  name: 'sourcersgmodule'
  params: {
    name: sourcersg
    location: resourceGroupLocation
  }
}

module destinationRsg 'resourcegroup.bicep' = {
  name: 'destrsgmodule'
  params: {
    name: destrsg
    location: resourceGroupLocation
  }
}

module storageAcct 'storage.bicep' = {
  name: 'storageModule'
  scope: resourceGroup(sourcersg)
  params: {
    storageLocation: resourceGroupLocation
  }
  dependsOn: [sourceRsg]
}
