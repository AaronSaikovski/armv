/*
MIT License

Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
targetScope = 'subscription'

param sourcersg string = 'src-rsg'
param destrsg string = 'dest-rsg'
param resourceGroupLocation string = 'australiaeast'
param aciCount int = 2

module sourceRsg './modules/resourcegroup.bicep' = {
  name: sourcersg
  params: {
    name: sourcersg
    location: resourceGroupLocation
  }
}

module destinationRsg './modules/resourcegroup.bicep' = {
  name: destrsg
  params: {
    name: destrsg
    location: resourceGroupLocation
  }
}

module webApp './modules/webapp.bicep' = {
  name: 'webappmodule'
  scope: resourceGroup(sourceRsg.name)
}

module containerInstance 'modules/aci.bicep' = {
  name: 'acimodule'
  scope: resourceGroup(sourceRsg.name)
  params: {
    aciCount: aciCount
  }
}
// module storageAcct './modules/storage.bicep' = {
//   name:'storagemodule'
//   params: {

//       storageLocation:resourceGroupLocation
//   }
//   scope: resourceGroup(sourceRsg.name)
// }
