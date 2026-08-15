package clinic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryConstructors_ReturnDomainPorts(t *testing.T) {
	assert.NotNil(t, NewClinicRepository(nil))
	assert.NotNil(t, NewClinicHolidayRepository(nil))
	assert.NotNil(t, NewClinicSettingsRepository(nil))
	assert.NotNil(t, NewClosingSpecialPeriodRepository(nil))
	assert.NotNil(t, NewCompanyRepository(nil))
}
